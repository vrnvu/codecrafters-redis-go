package server

import (
	"net"
	"testing"
)

type TestListener struct{}

func (l *TestListener) Accept() (net.Conn, error) {
	return nil, nil
}
func (l *TestListener) Close() error {
	return nil
}
func (l *TestListener) Addr() net.Addr {
	return nil
}

func TestServer(t *testing.T) {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		t.Fatalf("Listen error: %v", err)
	}
	server := NewServer(listener)
	defer server.Close()
}
