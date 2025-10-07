package protocol

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFrame(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		in      []byte
		wantErr bool
	}{
		{"simple", []byte("+OK\r\n"), false},
		{"error", []byte("-ERR oops\r\n"), false},
		{"bulk", []byte("$3\r\nfoo\r\n"), false},
		{"bulk-null", []byte("$-1\r\n"), false},
		{"array", []byte("*2\r\n$4\r\nPING\r\n$3\r\nfoo\r\n"), false},
		{"bad-type", []byte("?\r\n"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := bufio.NewReader(bytes.NewReader(tc.in))
			_, err := ReadFrame(r)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestWriteFrame(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   Frame
		out  []byte
	}{
		{"simple", SimpleString{Value: "OK"}, []byte("+OK\r\n")},
		{"error", Error{Message: "oops"}, []byte("-ERR oops\r\n")},
		{"bulk", BulkString{Bytes: []byte("foo")}, []byte("$3\r\nfoo\r\n")},
		{"bulk-null", BulkNullString{}, []byte("$-1\r\n")},
		{"array", Array{Elems: []Frame{BulkString{Bytes: []byte("PING")}, BulkString{Bytes: []byte("foo")}}}, []byte("*2\r\n$4\r\nPING\r\n$3\r\nfoo\r\n")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			w := bufio.NewWriter(&buf)
			err := tc.in.Write(w)
			assert.NoError(t, err)
			assert.Equal(t, tc.out, buf.Bytes())
		})
	}
}
