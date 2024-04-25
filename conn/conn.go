package conn

import (
	"net"
)

type Conn struct {
	netConn net.Conn
	r       *Reader
	w       *Writer
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		netConn: conn,
		r:       NewReader(conn),
		w:       NewWriter(conn),
	}
}

func (conn *Conn) ReadCommand() (Command, error) {
	return conn.r.ReadCommand()
}

func (conn *Conn) WriteStatus(b []byte) error {
	if err := conn.w.WriteStatus(b); err != nil {
		return err
	}
	return conn.w.Flush()
}

func (conn *Conn) WriteString(s string) error {
	if err := conn.w.WriteBytes([]byte(s)); err != nil {
		return err
	}
	return conn.w.Flush()
}

func (conn *Conn) WriteArray(a [][]byte) error {
	if err := conn.w.WriteArray(a); err != nil {
		return err
	}
	return conn.w.Flush()
}

func (conn *Conn) Flush() error {
	return conn.w.Flush()
}

func (conn *Conn) Close() error {
	if err := conn.w.Flush(); err != nil {
		return err
	}
	return conn.netConn.Close()
}
