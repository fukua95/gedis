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
	CmdSet      = "set"
	CmdGet      = "get"
	CmdPing     = "ping"
	CmdEcho     = "echo"
	CmdInfo     = "info"
	CmdReplConf = "replconf"
	CmdPsync    = "psync"
)

const (
	OptionSetEx     = "px"
	OptionInfoRep   = "replication"
	OptionReplLPort = "listening-port"
	OptionReplCapa  = "capa"
)

const (
	// client 对大小写敏感, 需要全大写.
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
