package tcp_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/hugjobk/go-tcp"
)

const ServerAddr = "127.0.0.1:9000"

func serveTCP(w io.Writer, p tcp.Packet) {
	fmt.Fprintf(w, "message %s->%s: %s", p.RemoteAddr, p.LocalAddr, p.Data)
}

func TestServer_ListenAndServe(t *testing.T) {
	s := tcp.Server{
		Addr:    ServerAddr,
		Handler: tcp.HandlerFunc(serveTCP),
	}
	t.Logf("Start listening at: %s", ServerAddr)
	if err := s.ListenAndServe(); err != nil {
		t.Fatal(err)
	}
}
