package command

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		in      string
		want    Command
		wantErr bool
	}{
		{"empty", "", nil, true},
		{"ping-upper", "PING", PingCommand{}, false},
		{"ping-lower", "ping", PingCommand{}, false},
		{"echo-upper", "ECHO foo", EchoCommand{Message: "foo"}, false},
		{"echo-lower", "echo foo", EchoCommand{Message: "foo"}, false},
		{"unknown", "unknown", nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			args := strings.Fields(tc.in)
			got, err := Parse(args)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tc.want.Type(), got.Type())
		})
	}
}
