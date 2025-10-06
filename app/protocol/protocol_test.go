package protocol

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newReader(t *testing.T, s string) *bufio.Reader {
	t.Helper()
	return bufio.NewReader(strings.NewReader(s))
}

func readN(t *testing.T, r io.Reader, n int) []byte {
	t.Helper()
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	assert.NoError(t, err)
	return buf
}

func TestReadLine(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"ok", "hello\r\n", "hello", false},
		{"no-crlf", "hello\n", "", true},
		{"empty", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := newReader(t, tc.in)
			got, err := ReadLine(r)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestReadBulk(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"ok", "$3\r\nfoo\r\n", "foo", false},
		{"empty-bulk", "$0\r\n\r\n", "", false},
		{"bad-prefix", "3\r\nfoo\r\n", "", true},
		{"bad-length", "$x\r\n", "", true},
		{"negative-length", "$-1\r\n", "", true},
		{"short-data", "$3\r\nfo\r\n", "", true},
		{"missing-crlf", "$3\r\nfooXX", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := newReader(t, tc.in)
			got, err := ReadBulk(r)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestReadArray(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		in      string
		want    []string
		wantErr bool
	}{
		{
			name: "simple-two",
			in:   "*2\r\n$4\r\nPING\r\n$3\r\nfoo\r\n",
			want: []string{"PING", "foo"},
		},
		{
			name: "empty-array",
			in:   "*0\r\n",
			want: []string{},
		},
		{
			name:    "bad-header",
			in:      "2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
			wantErr: true,
		},
		{
			name:    "bad-length",
			in:      "*$\r\n",
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := newReader(t, tc.in)
			got, err := ReadArray(r)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWriteSimpleString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		msg  string
		want []byte
	}{
		{"ok", "PONG", []byte("+PONG\r\n")},
		{"empty", "", []byte("+\r\n")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c1, c2 := net.Pipe()
			defer c1.Close()
			defer c2.Close()
			go WriteSimpleString(c1, tc.msg)
			buf := readN(t, c2, len(tc.want))
			assert.True(t, bytes.Equal(buf, tc.want), "got %q want %q", string(buf), string(tc.want))
		})
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		msg  string
		want []byte
	}{
		{"ok", "wrong number of arguments", []byte("-ERR wrong number of arguments\r\n")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c1, c2 := net.Pipe()
			defer c1.Close()
			defer c2.Close()
			go WriteError(c1, tc.msg)
			buf := readN(t, c2, len(tc.want))
			assert.True(t, bytes.Equal(buf, tc.want), "got %q want %q", string(buf), string(tc.want))
		})
	}
}
