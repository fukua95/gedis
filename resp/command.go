package resp

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
)

const (
	OptionSetEx     = "px"
	OptionInfoRep   = "replication"
	OptionReplLPort = "listening-port"
	OptionReplCapa  = "capa"
	OptionGetAck    = "GETACK"
	OptionAck       = "ACK"
	OptionDir       = "dir"
	OptionDBFile    = "dbfilename"
)

const (
	ReplyFullResync = "FULLRESYNC"
)

/*
var OptionArgMap = map[string]int{
	OptionPsync: 2,
}

func OptionArgCount(op string) int {
	if v, has := OptionArgMap[op]; has {
		return v
	}
	return 1
}
*/
