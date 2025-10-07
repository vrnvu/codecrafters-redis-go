package server

import (
	"net"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/store"
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
	testListener := &TestListener{}
	server := NewServer(testListener, store.NewStore())
	defer server.Close()
}
