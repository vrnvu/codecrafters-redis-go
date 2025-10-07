package command

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type PingCommand struct{}

func (c *PingCommand) Execute() protocol.SimpleString {
	return protocol.SimpleString{Value: "PONG"}
}

type EchoCommand struct {
	Message string
}

func (c *EchoCommand) Execute() protocol.SimpleString {
	return protocol.SimpleString{Value: c.Message}
}

type SetCommand struct {
	Key   string
	Value string
}

func (c *SetCommand) Execute() protocol.SimpleString {
	return protocol.SimpleString{Value: "OK"}
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
	case "SET":
		if len(arr.Elems) < 2 {
			return nil, fmt.Errorf("set command requires 1 argument")
		}
		key, ok := arr.Elems[1].(protocol.BulkString)
		if !ok {
			return nil, fmt.Errorf("set key must be a bulk string")
		}
		value, ok := arr.Elems[2].(protocol.BulkString)
		if !ok {
			return nil, fmt.Errorf("set value must be a bulk string")
		}
		return SetCommand{Key: string(key.Bytes), Value: string(value.Bytes)}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}
