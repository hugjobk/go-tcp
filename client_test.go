package tcp_test

import (
	"testing"
	"time"

	"github.com/hugjobk/go-tcp"
)

func TestClient_Write(t *testing.T) {
	cli := tcp.Client{RemoteAddr: ServerAddr, MaxConnCount: 10, DialTimeout: 1 * time.Second}
	if err := cli.Connect(); err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := cli.Write(time.Now().Add(1*time.Second), []byte("Hello World")); err != nil {
			t.Error(err)
		} else {
			t.Log("OK")
		}
		time.Sleep(1 * time.Second)
	}
}

func TestClient_WriteRead(t *testing.T) {
	cli := tcp.Client{RemoteAddr: ServerAddr, MaxConnCount: 10, DialTimeout: 1 * time.Second}
	if err := cli.Connect(); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 1024)
	for {
		if n, err := cli.WriteRead(time.Now().Add(1*time.Second), []byte("Hello World"), buf); err != nil {
			t.Error(err)
		} else {
			t.Log(string(buf[:n]))
		}
		time.Sleep(1 * time.Second)
	}
}

func TestClient_Ping(t *testing.T) {
	cli := tcp.Client{RemoteAddr: ServerAddr, MaxConnCount: 10, DialTimeout: 1 * time.Second}
	if err := cli.Connect(); err != nil {
		t.Fatal(err)
	}
	cli.Ping([]byte("ping"), 10, 1*time.Second, 1*time.Second)
	cli.Close()
	t.Log("disconnected from server")
}
