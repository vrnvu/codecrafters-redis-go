package command

import (
	"fmt"
	"strings"
)

type CommandType int

const (
	Ping CommandType = iota
)

type Command interface {
	Type() CommandType
}

type PingCommand struct{}

func (PingCommand) Type() CommandType { return Ping }

func Parse(args []string) (Command, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	switch strings.ToUpper(args[0]) {
	case "PING":
		return PingCommand{}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", args[0])
	}
}
