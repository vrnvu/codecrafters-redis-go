package command

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/rdb"
	"github.com/codecrafters-io/redis-starter-go/app/store"
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

func (c *SetCommand) Execute(store *store.Store) protocol.SimpleString {
	store.Set(c.Key, c.Value, nil)
	return protocol.SimpleString{Value: "OK"}
}

type SetTTLCommand struct {
	Key   string
	Value string
	TTL   time.Duration
}

func (c *SetTTLCommand) Execute(store *store.Store) protocol.SimpleString {
	store.Set(c.Key, c.Value, &c.TTL)
	return protocol.SimpleString{Value: "OK"}
}

type GetCommand struct {
	Key string
}

func (c *GetCommand) Execute(store *store.Store) protocol.Frame {
	value, ok := store.Get(c.Key)
	if !ok {
		return protocol.BulkNullString{}
	}

	return protocol.SimpleString{Value: value}
}

type IncrCommand struct {
	Key string
}

func (c *IncrCommand) Execute(store *store.Store) protocol.Frame {
	value, err := store.Incr(c.Key)
	if err != nil {
		return protocol.Error{Message: err.Error()}
	}

	return protocol.Integer{Value: value}
}

type ExecCommand struct{}

type DiscardCommand struct{}

type MultiCommand struct {
	Commands []any
}

func (c *MultiCommand) Execute(reader *bufio.Reader, writer *bufio.Writer, store *store.Store, file *rdb.File) protocol.Frame {
	protocol.SimpleString{Value: "OK"}.Write(writer)

read:
	for {
		frame, err := protocol.ReadFrame(reader)
		if err != nil {
			return protocol.Error{Message: err.Error()}
		}

		cmd, err := FromArray(frame.(protocol.Array))
		if err != nil {
			return protocol.Error{Message: err.Error()}
		}

		switch cmd.(type) {
		case MultiCommand:
			return protocol.Error{Message: "nested multi commands are not allowed"}
		case ExecCommand:
			break read
		case DiscardCommand:
			return protocol.SimpleString{Value: "OK"}
		default:
			c.Commands = append(c.Commands, cmd)
			protocol.SimpleString{Value: "QUEUED"}.Write(writer)
		}
	}

	if len(c.Commands) == 0 {
		return protocol.Array{Elems: []protocol.Frame{}}
	}

	results := make([]protocol.Frame, len(c.Commands))
	for i, cmd := range c.Commands {
		switch c := cmd.(type) {
		case PingCommand:
			results[i] = c.Execute()
		case EchoCommand:
			results[i] = c.Execute()
		case SetCommand:
			results[i] = c.Execute(store)
		case SetTTLCommand:
			results[i] = c.Execute(store)
		case GetCommand:
			results[i] = c.Execute(store)
		case IncrCommand:
			results[i] = c.Execute(store)
		case ExecCommand:
			msg := protocol.Error{Message: "EXEC without MULTI"}
			results[i] = msg
		case MultiCommand:
			msg := protocol.Error{Message: "nested multi commands are not allowed"}
			results[i] = msg
		case ConfigCommand:
			results[i] = c.Execute(file)
		default:
			results[i] = protocol.Error{Message: "unknown command"}
		}
	}

	return protocol.Array{Elems: results}
}

type ConfigCommand struct {
	Config string
}

func (c *ConfigCommand) Execute(file *rdb.File) protocol.Frame {
	switch c.Config {
	case "dbfilename":
		return protocol.Array{Elems: []protocol.Frame{protocol.BulkString{Bytes: []byte("dbfilename")}, protocol.BulkString{Bytes: []byte(file.DBFilename)}}}
	case "dir":
		return protocol.Array{Elems: []protocol.Frame{protocol.BulkString{Bytes: []byte("dir")}, protocol.BulkString{Bytes: []byte(file.Dir)}}}
	default:
		return protocol.Error{Message: fmt.Sprintf("unknown config: %s", c.Config)}
	}
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
		if len(arr.Elems) < 3 {
			return nil, fmt.Errorf("set command requires at least 2 arguments")
		}

		key, ok := arr.Elems[1].(protocol.BulkString)
		if !ok {
			return nil, fmt.Errorf("set key must be a bulk string")
		}

		value, ok := arr.Elems[2].(protocol.BulkString)
		if !ok {
			return nil, fmt.Errorf("set value must be a bulk string")
		}

		switch len(arr.Elems) {
		case 3:
			return SetCommand{Key: string(key.Bytes), Value: string(value.Bytes)}, nil
		case 5:
			unit, ok := arr.Elems[3].(protocol.BulkString)
			if !ok {
				return nil, fmt.Errorf("set expiration unit must be a bulk string")
			}

			ttlStr, ok := arr.Elems[4].(protocol.BulkString)
			if !ok {
				return nil, fmt.Errorf("set expiration value must be a bulk string")
			}

			ttlValue, err := strconv.ParseUint(string(ttlStr.Bytes), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid expiration value: %s", string(ttlStr.Bytes))
			}

			if ttlValue == 0 {
				return nil, fmt.Errorf("invalid expiration value: %d", ttlValue)
			}

			var ttl time.Duration
			switch string(unit.Bytes) {
			case "EX":
				ttl = time.Duration(ttlValue) * time.Second
			case "PX":
				ttl = time.Duration(ttlValue) * time.Millisecond
			default:
				return nil, fmt.Errorf("invalid expiration unit: %s", string(unit.Bytes))
			}

			return SetTTLCommand{Key: string(key.Bytes), Value: string(value.Bytes), TTL: ttl}, nil
		default:
			return nil, fmt.Errorf("set command requires 3 or 5 arguments")
		}
	case "GET":
		if len(arr.Elems) < 2 {
			return nil, fmt.Errorf("get command requires 1 argument")
		}
		key, ok := arr.Elems[1].(protocol.BulkString)
		if !ok {
			return nil, fmt.Errorf("get key must be a bulk string")
		}
		return GetCommand{Key: string(key.Bytes)}, nil
	case "INCR":
		if len(arr.Elems) < 2 {
			return nil, fmt.Errorf("incr command requires 1 argument")
		}
		key, ok := arr.Elems[1].(protocol.BulkString)
		if !ok {
			return nil, fmt.Errorf("incr key must be a bulk string")
		}
		return IncrCommand{Key: string(key.Bytes)}, nil
	case "MULTI":
		return MultiCommand{Commands: []any{}}, nil
	case "EXEC":
		return ExecCommand{}, nil
	case "DISCARD":
		return DiscardCommand{}, nil
	case "CONFIG":
		if len(arr.Elems) < 2 {
			return nil, fmt.Errorf("config command requires 1 argument")
		}
		config, ok := arr.Elems[1].(protocol.BulkString)
		if !ok {
			return nil, fmt.Errorf("config argument must be a bulk string")
		}
		switch string(config.Bytes) {
		case "dir":
			return ConfigCommand{Config: string(config.Bytes)}, nil
		case "dbfilename":
			return ConfigCommand{Config: string(config.Bytes)}, nil
		default:
			return nil, fmt.Errorf("unknown config: %s", string(config.Bytes))
		}
	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}
