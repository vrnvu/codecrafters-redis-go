package main

import (
	"bufio"
	"io"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		args, err := protocol.ReadArray(reader)
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

		switch cmd.(type) {
		case command.PingCommand:
			protocol.WriteSimpleString(conn, "PONG")
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
