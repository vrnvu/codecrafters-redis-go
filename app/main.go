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

func handleConnection(conn net.Conn) {
	defer conn.Close()

	r := readerPool.Get().(*bufio.Reader)
	r.Reset(conn)
	defer func() {
		r.Reset(nil)
		readerPool.Put(r)
	}()

	for {
		args, err := protocol.ReadArray(r)
		if err != nil {
			if err == io.EOF {
				return
			}

			protocol.WriteError(conn, err.Error())
			return
		}

		cmd, err := command.Parse(args)
		if err != nil {
			protocol.WriteError(conn, err.Error())
			continue
		}

		switch cmd := cmd.(type) {
		case command.PingCommand:
			protocol.WriteSimpleString(conn, "PONG")
		case command.EchoCommand:
			protocol.WriteSimpleString(conn, cmd.Message)
		default:
			protocol.WriteError(conn, "unknown cmd")
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
