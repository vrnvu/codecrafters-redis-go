package main

import (
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func main() {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}

	server := server.NewServer(listener)
	defer server.Close()

	log.Println("Listening on :6379")
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		go server.HandleConnection(conn)
	}
}
