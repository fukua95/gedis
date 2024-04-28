package server

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
	At(pos int) []byte
	SearchOption(op string) ([]byte, bool)
	SetArgs(b [][]byte)
	Result() interface{}
	SetResult(res interface{})
}

type command struct {
	args   [][]byte
	result any
}

func (cmd *command) Name() string {
	if len(cmd.args) == 0 {
		return ""
	}
	return strings.ToLower(string(cmd.At(0)))
}

func (cmd *command) Args() [][]byte {
	return cmd.args
}

func (cmd *command) At(pos int) []byte {
	if pos < 0 || pos >= len(cmd.args) {
		panic(ErrIndexOutOfRange)
	}
	return cmd.args[pos]
}

// don't use.
func (cmd *command) IntArg(pos int) (int, error) {
	if pos < 0 || pos >= len(cmd.args) {
		return 0, ErrIndexOutOfRange
	}
	arg := cmd.args[pos]
	return strconv.Atoi(string(arg))
}

func (cmd *command) SearchOption(op string) ([]byte, bool) {
	for i := 3; i < len(cmd.args); i++ {
		if strings.ToLower(string(cmd.args[i])) == op && i+1 < len(cmd.args) {
			return cmd.args[i+1], true
		}
	}
	return nil, false
}

func (cmd *command) SetArgs(b [][]byte) {
	cmd.args = b
}

func (cmd *command) Result() interface{} {
	return cmd.result
}

func (cmd *command) SetResult(r interface{}) {
	cmd.result = r
}
