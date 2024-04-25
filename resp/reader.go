package resp

import (
	"bufio"
	"fmt"
	"io"

	"github.com/codecrafters-io/redis-starter-go/util"
)

type Reader struct {
	r bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: *bufio.NewReader(r),
	}
}

// resp: https://redis.io/docs/latest/develop/reference/protocol-spec/#sending-commands-to-a-redis-server
func (r *Reader) ReadCommand() (Command, error) {
	cmd := new(command)
	len, err := r.readArrayLen()
	if err != nil {
		return nil, err
	}
	if len == 0 {
		return nil, fmt.Errorf("invalid command, size=0")
	}

	cmd.args = make([][]byte, len)
	for i := 0; i < len; i++ {
		arg, err := r.readArg()
		if err != nil {
			return nil, err
		}
		cmd.args[i] = arg
	}
	return cmd, nil
}

func (r *Reader) readArrayLen() (int, error) {
	line, err := r.readLine()
	if err != nil {
		return 0, err
	}
	if line[0] != RespArray {
		return 0, fmt.Errorf("redis: invalid command: %q", line)
	}
	return util.Atoi(line[1:])
}

func (r *Reader) readArg() ([]byte, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	// we only need to handle `RespString` currently.
	switch line[0] {
	case RespString:
		s, err := r.readLine()
		if err != nil {
			return nil, err
		}
		return util.BytesToLower(s), nil
	default:
		return nil, fmt.Errorf("redis: invalid command: %q", line)
	}
}

// readLine returns an error if:
// - there is a pending read error
// - or line does not end with \r\n
func (r *Reader) readLine() ([]byte, error) {
	b, err := r.r.ReadSlice('\n')
	if err != nil {
		if err != bufio.ErrBufferFull {
			return nil, err
		}

		full := make([]byte, len(b))
		copy(full, b)

		if b, err = r.r.ReadBytes('\n'); err != nil {
			return nil, err
		}

		full = append(full, b...)
		b = full
	}
	if len(b) <= 2 || b[len(b)-1] != '\n' || b[len(b)-2] != '\r' {
		return nil, fmt.Errorf("redis: invalid command: %q", b)
	}
	return b[:len(b)-2], nil
}
