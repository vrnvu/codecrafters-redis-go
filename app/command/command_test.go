package command

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/stretchr/testify/assert"
)

func TestFromArray(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		in      protocol.Array
		want    any
		wantErr bool
	}{
		{
			name: "ping",
			in: protocol.Array{
				Elems: []protocol.Frame{protocol.BulkString{Bytes: []byte("PING")}},
			},
			want: PingCommand{},
		},
		{
			name: "echo",
			in: protocol.Array{
				Elems: []protocol.Frame{
					protocol.BulkString{Bytes: []byte("ECHO")},
					protocol.BulkString{Bytes: []byte("hello")},
				},
			},
			want: EchoCommand{Message: "hello"},
		},
		{
			name:    "empty",
			in:      protocol.Array{Elems: []protocol.Frame{}},
			wantErr: true,
		},
		{
			name:    "null",
			in:      protocol.Array{Null: true},
			wantErr: true,
		},
		{
			name: "echo-missing-arg",
			in: protocol.Array{
				Elems: []protocol.Frame{protocol.BulkString{Bytes: []byte("ECHO")}},
			},
			wantErr: true,
		},
		{
			name: "unknown",
			in: protocol.Array{
				Elems: []protocol.Frame{protocol.BulkString{Bytes: []byte("UNKNOWN")}},
			},
			wantErr: true,
		},
		{
			name: "set",
			in: protocol.Array{
				Elems: []protocol.Frame{protocol.BulkString{Bytes: []byte("SET")}, protocol.BulkString{Bytes: []byte("key")}, protocol.BulkString{Bytes: []byte("value")}},
			},
			want: SetCommand{Key: "key", Value: "value"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := FromArray(tc.in)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
