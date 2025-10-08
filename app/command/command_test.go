package command

import (
	"testing"
	"time"

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
		{
			name: "set-ttl ex",
			in: protocol.Array{
				Elems: []protocol.Frame{protocol.BulkString{Bytes: []byte("SET")}, protocol.BulkString{Bytes: []byte("key")}, protocol.BulkString{Bytes: []byte("value")}, protocol.BulkString{Bytes: []byte("EX")}, protocol.BulkString{Bytes: []byte("10")}},
			},
			want: SetTTLCommand{Key: "key", Value: "value", TTL: 10 * time.Second},
		},
		{
			name: "set-ttl px",
			in: protocol.Array{
				Elems: []protocol.Frame{protocol.BulkString{Bytes: []byte("SET")}, protocol.BulkString{Bytes: []byte("key")}, protocol.BulkString{Bytes: []byte("value")}, protocol.BulkString{Bytes: []byte("PX")}, protocol.BulkString{Bytes: []byte("10")}},
			},
			want: SetTTLCommand{Key: "key", Value: "value", TTL: 10 * time.Millisecond},
		},
		{
			name: "set-ttl invalid",
			in: protocol.Array{
				Elems: []protocol.Frame{protocol.BulkString{Bytes: []byte("SET")}, protocol.BulkString{Bytes: []byte("key")}, protocol.BulkString{Bytes: []byte("value")}, protocol.BulkString{Bytes: []byte("INVALID")}, protocol.BulkString{Bytes: []byte("10")}},
			},
			wantErr: true,
		},
		{
			name: "set-ttl invalid value",
			in: protocol.Array{
				Elems: []protocol.Frame{protocol.BulkString{Bytes: []byte("SET")}, protocol.BulkString{Bytes: []byte("key")}, protocol.BulkString{Bytes: []byte("value")}, protocol.BulkString{Bytes: []byte("EX")}, protocol.BulkString{Bytes: []byte("0")}},
			},
			wantErr: true,
		},
		{
			name: "get",
			in: protocol.Array{
				Elems: []protocol.Frame{protocol.BulkString{Bytes: []byte("GET")}, protocol.BulkString{Bytes: []byte("key")}},
			},
			want: GetCommand{Key: "key"},
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
