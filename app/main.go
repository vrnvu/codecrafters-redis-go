package main

import (
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func main() {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}

	store := store.NewStore()

	server := server.NewServer(listener, store)
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
