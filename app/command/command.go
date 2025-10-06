package command

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type PingCommand struct{}
type EchoCommand struct {
	Message string
}

// FromArray converts a protocol.Array to a command
func FromArray(arr protocol.Array) (any, error) {
	if arr.Null || len(arr.Elems) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmdName, ok := arr.Elems[0].(protocol.BulkString)
	if !ok {
		return nil, fmt.Errorf("command must be a bulk string")
	}

	cmd := strings.ToUpper(string(cmdName.Bytes))
	switch cmd {
	case "PING":
		return PingCommand{}, nil
	case "ECHO":
		if len(arr.Elems) < 2 {
			return nil, fmt.Errorf("echo command requires 1 argument")
		}
		msg, ok := arr.Elems[1].(protocol.BulkString)
		if !ok {
			return nil, fmt.Errorf("echo argument must be a bulk string")
		}
		return EchoCommand{Message: string(msg.Bytes)}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}
