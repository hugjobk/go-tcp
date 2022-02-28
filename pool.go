package tcp

import (
	"crypto/tls"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type connPool struct {
	network      string
	remoteAddr   string
	maxConnCount int32
	dialTimeout  time.Duration
	tlsConfig    *tls.Config
	logger       *log.Logger

	connCount int32
	conns     chan *conn
	mtx       sync.RWMutex
	closed    bool
}

func (p *connPool) init() error {
	p.conns = make(chan *conn, int(p.maxConnCount))
	go func() {
		tempDelay := 5 * time.Millisecond //How long to sleep on dial failure
	loop:
		for {
			p.mtx.RLock()
			if p.closed {
				p.mtx.RUnlock()
				return
			}
			for atomic.LoadInt32(&p.connCount) < p.maxConnCount {
				if c, err := p.dial(); err != nil {
					p.mtx.RUnlock()
					p.logf("tcp: Dial error: %s; retrying in %s", err, tempDelay)
					time.Sleep(tempDelay)
					tempDelay *= 2
					if tempDelay > 1*time.Second {
						tempDelay = 1 * time.Second
					}
					continue loop
				} else {
					tempDelay = 5 * time.Millisecond
					p.conns <- &conn{Conn: c}
					atomic.AddInt32(&p.connCount, 1)
				}
			}
			p.mtx.RUnlock()
			time.Sleep(1 * time.Second)
		}
	}()
	return nil
}

func (p *connPool) get() <-chan *conn {
	return p.conns
}

func (p *connPool) put(c *conn) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	if p.closed || c.isClosed() {
		c.Close()
		atomic.AddInt32(&p.connCount, -1)
	} else {
		c.err = nil
		p.conns <- c
	}
}

func (p *connPool) dial() (net.Conn, error) {
	d := net.Dialer{Timeout: p.dialTimeout}
	if p.tlsConfig != nil {
		return tls.DialWithDialer(&d, p.network, p.remoteAddr, p.tlsConfig)
	}
	return d.Dial(p.network, p.remoteAddr)
}

func (p *connPool) close() error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.closed = true
	close(p.conns)
	for c := range p.conns {
		c.Close()
		atomic.AddInt32(&p.connCount, -1)
	}
	return nil
}

func (p *connPool) logf(format string, args ...interface{}) {
	p.logger.Printf(format, args...)
}
