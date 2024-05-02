package server

import (
	"fmt"
	"strings"
)

const (
	port       string = "port"
	replicaof  string = "replicaof"
	dbdir      string = "dir"
	dbfilename string = "dbfilename"
)

type Config struct {
	network    string
	port       string
	addr       string
	role       role
	masterAddr string
	dir        string
	dbfilename string
}

func NewConfig(args []string) *Config {
	conf := new(Config)
	for i := 1; i < len(args); i++ {
		arg := strings.ToLower(args[i])
		if strings.HasSuffix(arg, port) && i+1 < len(args) {
			conf.port = strings.ToLower(args[i+1])
		} else if strings.HasSuffix(arg, replicaof) && i+2 < len(args) {
			conf.role = roleReplica
			conf.masterAddr = fmt.Sprintf("%s:%s", strings.ToLower(args[i+1]), strings.ToLower(args[i+2]))
		} else if strings.HasSuffix(arg, dbdir) && i+1 < len(args) {
			conf.dir = args[i+1]
		} else if strings.HasSuffix(arg, dbfilename) && i+1 < len(args) {
			conf.dbfilename = args[i+1]
		}
	}

	if conf.network == "" {
		conf.network = "tcp"
	}
	if conf.port == "" {
		conf.port = "6379"
	}
	conf.addr = fmt.Sprintf("0.0.0.0:%s", conf.port)
	if conf.masterAddr == "" {
		conf.role = roleMaster
	}
	return conf
}
