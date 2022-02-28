package tcp

import (
	"net"
	"time"
)

type conn struct {
	net.Conn
	err error
}

func (c *conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	c.err = err
	return n, err
}

func (c *conn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	c.err = err
	return n, err
}

func (c *conn) SetDeadline(deadline time.Time) error {
	err := c.Conn.SetDeadline(deadline)
	c.err = err
	return err
}

func (c *conn) SetReadDeadline(deadline time.Time) error {
	err := c.Conn.SetReadDeadline(deadline)
	c.err = err
	return err
}

func (c *conn) SetWriteDeadline(deadline time.Time) error {
	err := c.Conn.SetWriteDeadline(deadline)
	c.err = err
	return err
}

func (c *conn) isClosed() bool {
	if c.err == nil {
		return false
	}
	if ne, ok := c.err.(net.Error); ok && ne.Temporary() {
		return false
	}
	return true
}
