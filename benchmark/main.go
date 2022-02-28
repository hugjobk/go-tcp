package main

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/hugjobk/go-benchmark"
	"github.com/hugjobk/go-tcp"
)

const ServerAddr = "127.0.0.1:9001"

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024)
	},
}

func initServer() {
	srv := tcp.Server{
		Addr: ServerAddr,
		Handler: tcp.HandlerFunc(func(w io.Writer, p tcp.Packet) {
			w.Write([]byte("World"))
		}),
	}
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

func main() {
	go initServer()
	cli := tcp.Client{RemoteAddr: ServerAddr, MaxConnCount: 100, DialTimeout: 1 * time.Second}
	if err := cli.Connect(); err != nil {
		panic(err)
	}
	b := benchmark.Benchmark{
		WorkerCount:  100,
		Duration:     30 * time.Second,
		LatencyStart: 5 * time.Millisecond,
		LatencyStep:  6,
		ShowProcess:  true,
	}
	b.Run("Benchmark WriteRead", func(int) error {
		buf := bufferPool.Get().([]byte)
		_, err := cli.WriteRead(time.Now().Add(1*time.Second), []byte("Hello"), buf)
		bufferPool.Put(buf)
		return err
	}).Report(os.Stdout).PrintErrors(os.Stdout, 10)
}
