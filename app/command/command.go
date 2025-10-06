package command

import (
	"fmt"
	"strings"
)

type Command interface {
	Name() string
}

type PingCommand struct{}

func (PingCommand) Name() string { return "PING" }

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
