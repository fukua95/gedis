package server

type role string

const (
	master role = "master"
	slave  role = "slave"
)

type replication struct {
	role       role
	masterAddr string
}

func newReplication(isSlave bool, masterAddr string) *replication {
	role := master
	if isSlave {
		role = slave
	}
	return &replication{
		role:       role,
		masterAddr: masterAddr,
	}
}
