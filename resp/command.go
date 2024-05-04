package resp

import "errors"

// redis resp protocol data type.
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
)
