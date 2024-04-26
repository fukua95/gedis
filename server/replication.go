package server

import "github.com/codecrafters-io/redis-starter-go/util"

type role string

const (
	roleMaster role = "master"
	roleSlave  role = "slave"
)

type replication struct {
	role             role
	masterAddr       string
	masterReplID     string
	masterReplOffset int
}

func newReplication(isSlave bool, masterAddr string) *replication {
	role := roleMaster
	if isSlave {
		role = roleSlave
	}
	repl := &replication{
		role:       role,
		masterAddr: masterAddr,
	}
	if repl.role == roleMaster {
		repl.masterReplID = util.RandomAlphanumericString(40)
		repl.masterReplOffset = 0
	}
	return repl
}
