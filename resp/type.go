package resp

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/util"
)

func String(s string) []byte {
	return []byte(fmt.Sprintf("%c%s\r\n%s\r\n", RespString, util.Itoa(len(s)), s))
}

func ArrayHeader(l int) []byte {
	return []byte(fmt.Sprintf("%c%s\r\n", RespArray, util.Itoa(l)))
}

func Status(s string) []byte {
	return []byte(fmt.Sprintf("%c%s\r\n", RespStatus, s))
}

func Error(e string) []byte {
	return []byte(fmt.Sprintf("%cERR %s\r\n", RespError, e))
}

func NilString() []byte {
	return []byte(fmt.Sprintf("%c-1\r\n", RespString))
}

// without tail `\r\n`
func RdbContent(content []byte) []byte {
	return []byte(fmt.Sprintf("%c%s\r\n%s", RespString, util.Itoa(len(content)), content))
}

func Integer(v int) []byte {
	return []byte(fmt.Sprintf("%c%s\r\n", RespInt, util.Itoa(v)))
}
