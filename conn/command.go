package conn

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
