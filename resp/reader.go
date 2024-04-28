package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/util"
)

var (
	ErrInvalidReply = errors.New("invalid reply")
)

type Reader struct {
	rd *bufio.Reader
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{
		rd: bufio.NewReader(rd),
	}
}

func (r *Reader) ReadReply() (interface{}, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}

	switch line[0] {
	case RespStatus:
		return string(line[1:]), nil
	case RespInt:
		return strconv.ParseInt(string(line[1:]), 10, 64)
	case RespFloat:
		return r.float(line)
	case RespBool:
		return r.bool(line)
	case RespString:
		return r.readStringContent(line)
		//case RespArray, RespSet, RespPush:
		//	return r.readSlice(line)
		//case RespMap:
		//	return r.readMap(line)
	}
	return nil, fmt.Errorf("redis: can't parse %.100q", line)
}

// readLine returns an error if:
// - there is a pending read error;
// - or line does not end with \r\n
func (r *Reader) readLine() ([]byte, error) {
	b, err := r.rd.ReadSlice('\n')
	if err != nil {
		if err != bufio.ErrBufferFull {
			return nil, err
		}

		full := make([]byte, len(b))
		copy(full, b)
		b, err = r.rd.ReadBytes('\n')
		if err != nil {
			return nil, err
		}

		full = append(full, b...)
		b = full
	}
	if len(b) <= 2 || b[len(b)-1] != '\n' || b[len(b)-2] != '\r' {
		return nil, fmt.Errorf("redis: invalid reply: %q", b)
	}
	return b[:len(b)-2], nil
}

func (r *Reader) float(line []byte) (float64, error) {
	v := string(line[1:])
	switch v {
	case "inf":
		return math.Inf(1), nil
	case "-inf":
		return math.Inf(-1), nil
	case "nan", "-nan":
		return math.NaN(), nil
	}
	return strconv.ParseFloat(v, 64)
}

func (r *Reader) bool(line []byte) (bool, error) {
	switch string(line[1:]) {
	case "t":
		return true, nil
	case "f":
		return false, nil
	}
	return false, fmt.Errorf("redis: can't parse bool reply: %q", line)
}

func (r *Reader) readStringContent(line []byte) (string, error) {
	n, err := parseLen(line)
	if err != nil {
		return "", err
	}

	b := make([]byte, n+2)
	_, err = io.ReadFull(r.rd, b)
	if err != nil {
		return "", err
	}
	return string(b[:n]), nil
}

func (r *Reader) ReadInt() (int64, error) {
	line, err := r.readLine()
	if err != nil {
		return 0, err
	}
	switch line[0] {
	case RespInt, RespStatus:
		return strconv.ParseInt(string(line[1:]), 10, 64)
	case RespString:
		s, err := r.readStringContent(line)
		if err != nil {
			return 0, err
		}
		return strconv.ParseInt(s, 10, 64)
	}
	return 0, fmt.Errorf("redis: can't parse int reply: %.100q", line)
}

func (r *Reader) ReadString() (string, error) {
	line, err := r.readLine()
	if err != nil {
		return "", err
	}
	switch line[0] {
	case RespStatus, RespInt, RespFloat:
		return string(line[1:]), nil
	case RespString:
		return r.readStringContent(line)
	}
	return "", fmt.Errorf("redis: can't parse reply=%.100q reading string", line)
}

func (r *Reader) ReadSlice() ([][]byte, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	return r.readSliceContent(line)
}

func (r *Reader) readSliceContent(line []byte) ([][]byte, error) {
	n, err := parseLen(line)
	if err != nil {
		return nil, err
	}

	val := make([][]byte, n)
	for i := 0; i < n; i++ {
		v, err := r.ReadString()
		if err != nil {
			return nil, err
		}
		val[i] = []byte(v)
	}
	return val, nil
}

func parseLen(line []byte) (n int, err error) {
	n, err = util.Atoi(line[1:])
	if err != nil {
		return 0, err
	}
	if n < -1 {
		return 0, fmt.Errorf("redis: invalid reply: %q", line)
	}
	switch line[0] {
	case RespString, RespVerbatim, RespArray, RespSet, RespPush, RespMap, RespAttr:
		if n == -1 {
			return 0, nil
		}
	}
	return n, nil
}
