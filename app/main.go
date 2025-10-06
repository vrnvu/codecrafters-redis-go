package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

var readerPool = sync.Pool{
	New: func() any { return bufio.NewReaderSize(nil, 4096) },
}

var writerPool = sync.Pool{
	New: func() any { return bufio.NewWriterSize(nil, 4096) },
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	r := readerPool.Get().(*bufio.Reader)
	r.Reset(conn)
	defer func() {
		r.Reset(nil)
		readerPool.Put(r)
	}()

	w := writerPool.Get().(*bufio.Writer)
	w.Reset(conn)
	defer func() {
		w.Flush()
		w.Reset(nil)
		writerPool.Put(w)
	}()

	for {
		frame, err := protocol.ReadFrame(r)
		if err != nil {
			if err == io.EOF {
				return
			}
			_ = protocol.WriteFrame(w, protocol.Error{Message: err.Error()})
			return
		}

		request, ok := frame.(protocol.Array)
		if !ok || request.Null || len(request.Elems) == 0 {
			_ = protocol.WriteFrame(w, protocol.Error{Message: "invalid request"})
			continue
		}

		// Convert protocol array to command
		cmd, err := command.FromArray(request)
		if err != nil {
			_ = protocol.WriteFrame(w, protocol.Error{Message: err.Error()})
			continue
		}

		// Execute command
		switch c := cmd.(type) {
		case command.PingCommand:
			_ = protocol.WriteFrame(w, protocol.SimpleString{Value: "PONG"})
		case command.EchoCommand:
			_ = protocol.WriteFrame(w, protocol.SimpleString{Value: c.Message})
		default:
			_ = protocol.WriteFrame(w, protocol.Error{Message: "unknown command"})
		}
	}

}

func main() {
	ln, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}
	log.Println("Listening on :6379")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}
