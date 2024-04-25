package resp

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrIndexOutOfRange = errors.New("redis: unexpected index")
)

type Command interface {
	Name() string
	Args() [][]byte
	StringArg(pos int) (string, error)
	OptionSetEx() ([]byte, bool)
}

type command struct {
	args [][]byte
}

func (cmd *command) Name() string {
	if len(cmd.args) == 0 {
		return ""
	}
	name, err := cmd.StringArg(0)
	if err != nil {
		return ""
	}
	return strings.ToLower(name)
}

func (cmd *command) Args() [][]byte {
	return cmd.args
}

func (cmd *command) StringArg(pos int) (string, error) {
	if pos < 0 || pos >= len(cmd.args) {
		return "", ErrIndexOutOfRange
	}
	arg := cmd.args[pos]
	return string(arg), nil
}

// don't use.
func (cmd *command) IntArg(pos int) (int, error) {
	if pos < 0 || pos >= len(cmd.args) {
		return 0, ErrIndexOutOfRange
	}
	arg := cmd.args[pos]
	return strconv.Atoi(string(arg))
}

func (cmd *command) OptionSetEx() ([]byte, bool) {
	for i := 3; i < len(cmd.args); i++ {
		if strings.ToLower(string(cmd.args[i])) == OptionSetEx && i+1 < len(cmd.args) {
			return cmd.args[i+1], true
		}
	}
	return nil, false
}
