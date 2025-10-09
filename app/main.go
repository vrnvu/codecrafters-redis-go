package main

import (
	"flag"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/rdb"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func main() {
	dirFlag := flag.String("dir", "/tmp/redis-data", "directory containing the RDB file")
	dbFilenameFlag := flag.String("dbfilename", "dump.rdb", "RDB filename")
	flag.Parse()

	file := rdb.NewFile(*dirFlag, *dbFilenameFlag)
	if err := file.Open(); err != nil {
		log.Fatalf("open rdb file: %v", err)
	}
	defer file.Close()

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}

	store := store.NewStore()

	server := server.NewServer(listener, store, file)
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
