package server

type role string

const (
	master role = "master"
)

type replication struct {
	role role
}

func newReplication(role role) *replication {
	return &replication{
		role: role,
	}
}
