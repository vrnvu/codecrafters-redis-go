package command

import (
	"fmt"
	"strings"
)

type CommandType int

const (
	Ping CommandType = iota
	Echo
)

type Command interface {
	Type() CommandType
}

type PingCommand struct{}

func (PingCommand) Type() CommandType { return Ping }

type EchoCommand struct {
	Message string
}

func (EchoCommand) Type() CommandType { return Echo }

func Parse(args []string) (Command, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	switch strings.ToUpper(args[0]) {
	case "PING":
		return PingCommand{}, nil
	case "ECHO":
		if len(args) < 2 {
			return nil, fmt.Errorf("echo command requires 1 argument")
		}
		return EchoCommand{Message: args[1]}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", args[0])
	}
}
