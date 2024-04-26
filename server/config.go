package server

import (
	"fmt"
	"strings"
)

const (
	port      string = "port"
	replicaof string = "replicaof"
)

type Config struct {
	network    string
	port       string
	addr       string
	isSlave    bool
	masterAddr string
}

func NewConfig(args []string) *Config {
	conf := new(Config)
	for i := 1; i < len(args); i++ {
		arg := strings.ToLower(args[i])
		if strings.HasSuffix(arg, port) && i+1 < len(args) {
			conf.port = strings.ToLower(args[i+1])
		}
		if strings.HasSuffix(arg, replicaof) && i+2 < len(args) {
			conf.isSlave = true
			conf.masterAddr = fmt.Sprintf("%s:%s", strings.ToLower(args[i+1]), strings.ToLower(args[i+2]))
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
		conf.isSlave = false
	}
	return conf
}
