package tcp

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Handler interface {
	ServeTCP(w io.Writer, p Packet)
}

type HandlerFunc func(w io.Writer, p Packet)

func (f HandlerFunc) ServeTCP(w io.Writer, p Packet) {
	f(w, p)
}

type Server struct {
	Network          string
	Addr             string
	TLSConfig        *tls.Config
	Handler          Handler
	Logger           *log.Logger
	PacketSplitFunc  PacketSplitFunc
	PacketUnwrapFunc PacketUnwrapFunc

	l     net.Listener
	mtx   sync.Mutex
	conns map[net.Conn]struct{}
}

func (srv *Server) ListenAndServe() error {
	if err := srv.init(); err != nil {
		return err
	}
	if err := srv.listen(); err != nil {
		return err
	}
	tempDelay := 5 * time.Millisecond //How long to sleep on accept failure
	for {
		c, err := srv.l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				srv.logf("tcp: Accept error: %s; retrying in %s", err, tempDelay)
				time.Sleep(tempDelay)
				tempDelay *= 2
				if tempDelay > 1*time.Second {
					tempDelay = 1 * time.Second
				}
				continue
			}
			return err
		}
		tempDelay = 5 * time.Millisecond
		srv.addConn(c)
		go srv.handleConn(c)
	}
}

func (srv *Server) Close() error {
	srv.l.Close()
	srv.mtx.Lock()
	for conn := range srv.conns {
		conn.Close()
	}
	srv.mtx.Unlock()
	return nil
}

func (srv *Server) init() error {
	if srv.Network == "" {
		srv.Network = "tcp"
	}
	if srv.Addr == "" {
		return errors.New("tcp: Server's Addr is empty")
	}
	if srv.Handler == nil {
		return errors.New("tcp: Server's Handler is nil")
	}
	if srv.Logger == nil {
		srv.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	if srv.PacketSplitFunc == nil {
		srv.PacketSplitFunc = LengthValuePacketSplitFunc(MagicNumber)
	}
	if srv.PacketUnwrapFunc == nil {
		srv.PacketUnwrapFunc = LengthValuePacketUnwrapFunc
	}
	srv.conns = make(map[net.Conn]struct{})
	return nil
}

func (srv *Server) listen() error {
	var err error
	if srv.TLSConfig == nil {
		srv.l, err = net.Listen(srv.Network, srv.Addr)
	} else {
		srv.l, err = tls.Listen(srv.Network, srv.Addr, srv.TLSConfig)
	}
	return err
}

func (srv *Server) handleConn(c net.Conn) {
	localAddr := c.LocalAddr().String()
	remoteAddr := c.RemoteAddr().String()
	srv.logf("tcp: New connection %s->%s", remoteAddr, localAddr)
	scanner := bufio.NewScanner(c)
	scanner.Split(bufio.SplitFunc(srv.PacketSplitFunc))
	for scanner.Scan() {
		packetData := srv.PacketUnwrapFunc(scanner.Bytes())
		srv.Handler.ServeTCP(c, Packet{Data: packetData, LocalAddr: localAddr, RemoteAddr: remoteAddr})
	}
	if err := scanner.Err(); err != nil {
		srv.logf("tcp: %s", err)
	}
	srv.removeConn(c)
	srv.logf("tcp: Closed connection %s->%s", remoteAddr, localAddr)
}

func (srv *Server) addConn(c net.Conn) {
	srv.mtx.Lock()
	srv.conns[c] = struct{}{}
	srv.mtx.Unlock()
}

func (srv *Server) removeConn(c net.Conn) {
	srv.mtx.Lock()
	delete(srv.conns, c)
	srv.mtx.Unlock()
}

func (srv *Server) logf(format string, args ...interface{}) {
	srv.Logger.Printf(format, args...)
}
