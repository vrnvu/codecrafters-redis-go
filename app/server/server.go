package server

import (
	"bufio"
	"io"
	"log"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type Server struct {
	listener   net.Listener
	readerPool *sync.Pool
	writerPool *sync.Pool
	store      *store.Store
}

func NewServer(listener net.Listener, store *store.Store) *Server {
	readerPool := sync.Pool{New: func() any { return bufio.NewReaderSize(nil, 4096) }}
	writerPool := sync.Pool{New: func() any { return bufio.NewWriterSize(nil, 4096) }}
	return &Server{listener: listener, readerPool: &readerPool, writerPool: &writerPool, store: store}
}

func (s *Server) Accept() (net.Conn, error) {
	return s.listener.Accept()
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	reader := s.readerPool.Get().(*bufio.Reader)
	reader.Reset(conn)
	defer func() {
		reader.Reset(nil)
		s.readerPool.Put(reader)
	}()

	writer := s.writerPool.Get().(*bufio.Writer)
	writer.Reset(conn)
	defer func() {
		writer.Reset(nil)
		s.writerPool.Put(writer)
	}()

	for {
		frame, err := protocol.ReadFrame(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			if err := (protocol.Error{Message: err.Error()}.Write(writer)); err != nil {
				log.Printf("writing error response: %v", err)
				return
			}

			return
		}

		request, ok := frame.(protocol.Array)
		if !ok || request.Null || len(request.Elems) == 0 {
			if err := (protocol.Error{Message: "invalid request"}.Write(writer)); err != nil {
				log.Printf("writing error response: %v", err)
				return
			}
			continue
		}

		cmd, err := command.FromArray(request)
		if err != nil {
			if err := (protocol.Error{Message: err.Error()}.Write(writer)); err != nil {
				log.Printf("writing error response: %v", err)
				return
			}
			return
		}

		switch c := cmd.(type) {
		case command.PingCommand:
			res := c.Execute()
			if err := res.Write(writer); err != nil {
				log.Printf("writing response: %v", err)
				return
			}
		case command.EchoCommand:
			res := c.Execute()
			if err := res.Write(writer); err != nil {
				log.Printf("writing response: %v", err)
				return
			}
		case command.SetCommand:
			res := c.Execute(s.store)
			if err := res.Write(writer); err != nil {
				log.Printf("writing response: %v", err)
				return
			}
		case command.SetTTLCommand:
			res := c.Execute(s.store)
			if err := res.Write(writer); err != nil {
				log.Printf("writing response: %v", err)
				return
			}
		case command.GetCommand:
			res := c.Execute(s.store)
			if err := res.Write(writer); err != nil {
				log.Printf("writing response: %v", err)
				return
			}
		case command.IncrCommand:
			res := c.Execute(s.store)
			if err := res.Write(writer); err != nil {
				log.Printf("writing response: %v", err)
				return
			}
		case command.ExecCommand:
			msg := protocol.Error{Message: "EXEC without MULTI"}
			if err := msg.Write(writer); err != nil {
				log.Printf("writing response: %v", err)
				return
			}
		case command.MultiCommand:
			res := c.Execute(reader, writer, s.store)
			if err := res.Write(writer); err != nil {
				log.Printf("writing response: %v", err)
				return
			}
		default:
			if err := (protocol.Error{Message: "unknown command"}.Write(writer)); err != nil {
				log.Printf("writing error response: %v", err)
				return
			}
			return
		}
	}

}
