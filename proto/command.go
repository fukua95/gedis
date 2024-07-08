package proto

import (
	"errors"
	"fmt"

	"github.com/fukua95/gedis/util"
)

// redis protocol data type.
const (
	RespStatus    = '+' // +<string>\r\n
	RespError     = '-' // -<string>\r\n
	RespString    = '$' // $<length>\r\n<bytes>\r\n
	RespInt       = ':' // :<number>\r\n
	RespNil       = '_' // _\r\n
	RespFloat     = ',' // ,<float>\r\n
	RespBool      = '#' // true: #t\r\n false: #f\r\n
	RespBlobError = '!' // !<length>\r\n<bytes>\r\n
	RespVerbatim  = '=' // =<length>\r\nFORMAT:<bytes>\r\n
	RespBigInt    = '(' // (<big number>\r\n
	RespArray     = '*' // //*<len>\r\n
	RespMap       = '%' // %<len>\r\n
	RespSet       = '~' // ~<len>\r\n
	RespAttr      = '|' // |<len>\r\n(key)\r\n(value)\r\n... + command reply
	RespPush      = '>' // ><len>\r\n
)

const (
	CmdSet      = "SET"
	CmdGet      = "GET"
	CmdPing     = "PING"
	CmdEcho     = "ECHO"
	CmdInfo     = "INFO"
	CmdReplConf = "REPLCONF"
	CmdPsync    = "PSYNC"
	CmdWait     = "WAIT"
	CmdConfig   = "CONFIG"
	CmdKeys     = "KEYS"
	CmdType     = "TYPE"
	CmdXAdd     = "XADD"
	CmdXRange   = "XRANGE"
	CmdXRead    = "XREAD"
)

const (
	OptionSetEx          = "px"
	OptionInfoRep        = "replication"
	OptionReplLPort      = "listening-port"
	OptionReplCapa       = "capa"
	OptionGetAck         = "GETACK"
	OptionAck            = "ACK"
	OptionDir            = "dir"
	OptionDBFile         = "dbfilename"
	OptionBlock          = "block"
	OptionStreamIDNewest = "$"
)

const (
	ReplyFullResync = "FULLRESYNC"
)

var (
	ErrStreamIDInvalid = errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
	ErrStreamIDIllegal = errors.New("The ID specified in XADD must be greater than 0-0")
	ErrInvalidCommand  = errors.New("invalid command")
	ErrInvalidReply    = errors.New("invalid reply")
)

func String(s string) []byte {
	return []byte(fmt.Sprintf("%c%s\r\n%s\r\n", RespString, util.Itoa(len(s)), s))
}

func ArrayHeader(l int) []byte {
	return []byte(fmt.Sprintf("%c%s\r\n", RespArray, util.Itoa(l)))
}

func Array(strs []string) []byte {
	arr := ArrayHeader(len(strs))
	for _, str := range strs {
		arr = append(arr, String(str)...)
	}
	return arr
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
