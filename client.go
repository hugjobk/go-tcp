package tcp

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"os"
	"time"
)

const PingRespMaxSize = 1024

type Client struct {
	Network        string
	RemoteAddr     string
	MaxConnCount   int32
	DialTimeout    time.Duration
	TLSConfig      *tls.Config
	Logger         *log.Logger
	PacketWrapFunc PacketWrapFunc

	pool *connPool
}

func (cli *Client) Connect() error {
	if err := cli.init(); err != nil {
		return err
	}
	cli.pool = &connPool{
		network:      cli.Network,
		remoteAddr:   cli.RemoteAddr,
		maxConnCount: cli.MaxConnCount,
		dialTimeout:  cli.DialTimeout,
		tlsConfig:    cli.TLSConfig,
		logger:       cli.Logger,
	}
	return cli.pool.init()
}

func (cli *Client) Write(deadline time.Time, data []byte) (int, error) {
	return cli.WriteRaw(deadline, cli.PacketWrapFunc(data))
}

func (cli *Client) WriteRaw(deadline time.Time, data []byte) (int, error) {
	stop := time.NewTimer(time.Until(deadline))
	defer stop.Stop()
	for {
		var (
			c  *conn
			ok bool
		)
		select {
		case c, ok = <-cli.pool.get():
			if !ok {
				return 0, ErrClosed
			}
		case <-stop.C:
			return 0, ErrTimeout
		}
		if err := c.SetWriteDeadline(deadline); err != nil {
			cli.pool.put(c)
			return 0, err
		}
		n, err := c.Write(data)
		cli.pool.put(c)
		if err != nil && !os.IsTimeout(err) {
			continue
		}
		return n, err
	}
}

func (cli *Client) WriteRead(deadline time.Time, data []byte, buf []byte) (int, error) {
	return cli.WriteReadRaw(deadline, cli.PacketWrapFunc(data), buf)
}

func (cli *Client) WriteReadRaw(deadline time.Time, data []byte, buf []byte) (int, error) {
	stop := time.NewTimer(time.Until(deadline))
	defer stop.Stop()
	for {
		var (
			c  *conn
			ok bool
		)
		select {
		case c, ok = <-cli.pool.get():
			if !ok {
				return 0, ErrClosed
			}
		case <-stop.C:
			return 0, ErrTimeout
		}
		if err := c.SetWriteDeadline(deadline); err != nil {
			cli.pool.put(c)
			return 0, err
		}
		if _, err := c.Write(data); err != nil {
			cli.pool.put(c)
			if !os.IsTimeout(err) {
				continue
			}
			return 0, err
		}
		n, err := c.Read(buf)
		cli.pool.put(c)
		if err == io.EOF {
			continue
		}
		return n, err
	}
}

func (cli *Client) Ping(data []byte, maxRetry int, interval time.Duration, timeout time.Duration) {
	cli.PingRaw(cli.PacketWrapFunc(data), maxRetry, interval, timeout)
}

func (cli *Client) PingRaw(data []byte, maxRetry int, interval time.Duration, timeout time.Duration) {
	buf := make([]byte, PingRespMaxSize)
	for retry := 0; retry <= maxRetry; {
		if _, err := cli.WriteReadRaw(time.Now().Add(timeout), data, buf); err != nil {
			cli.logf("tcp: Ping error: %s; retry=%d", err, retry)
			retry++
		} else {
			retry = 0
		}
		time.Sleep(interval)
	}
}

func (cli *Client) Close() error {
	return cli.pool.close()
}

func (cli *Client) init() error {
	if cli.Network == "" {
		cli.Network = "tcp"
	}
	if cli.RemoteAddr == "" {
		return errors.New("tcp: Client's RemoteAddr is empty")
	}
	if cli.MaxConnCount <= 0 {
		cli.MaxConnCount = 1
	}
	if cli.DialTimeout == 0 {
		cli.DialTimeout = 3 * time.Second
	}
	if cli.Logger == nil {
		cli.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	if cli.PacketWrapFunc == nil {
		cli.PacketWrapFunc = LengthValuePacketWrapFunc(MagicNumber)
	}
	return nil
}

func (cli *Client) logf(format string, args ...interface{}) {
	cli.Logger.Printf(format, args...)
}
