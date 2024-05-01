package server

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

type Conn struct {
	netConn net.Conn
	r       *resp.Reader
	w       *resp.Writer
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		netConn: conn,
		r:       resp.NewReader(conn),
		w:       resp.NewWriter(conn),
	}
}

func (conn *Conn) ReadCommand() (Command, error) {
	args, err := conn.r.ReadSlice()
	if err != nil {
		return nil, err
	}
	cmd := &command{args: args}
	return cmd, nil
}

// golang 目前不支持 struct's method type.
func (conn *Conn) ReadStatusReply() (string, error) {
	v, err := conn.r.ReadReply()
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (conn *Conn) ReadSliceReply() ([][]byte, error) {
	return conn.r.ReadSlice()
}

func (conn *Conn) ReadRdb() ([]byte, error) {
	s, err := conn.r.ReadRdb()
	return s, err
}

func (conn *Conn) WriteCommand(cmd Command) error {
	return conn.WriteSlice(cmd.Args())
}

func (conn *Conn) WriteStatus(b []byte) error {
	if err := conn.w.WriteStatus(b); err != nil {
		return err
	}
	return conn.w.Flush()
}

func (conn *Conn) WriteStatusOK() error {
	return conn.WriteStatus([]byte("OK"))
}

func (conn *Conn) WriteString(s string) error {
	if err := conn.w.WriteBytes([]byte(s)); err != nil {
		return err
	}
	return conn.w.Flush()
}

func (conn *Conn) WriteNilBulkString() error {
	if err := conn.w.WriteNilBulkString(); err != nil {
		return err
	}
	return conn.w.Flush()
}

func (conn *Conn) WriteSlice(a [][]byte) error {
	if err := conn.w.WriteSlice(a); err != nil {
		return err
	}
	return conn.w.Flush()
}

func (conn *Conn) WriteRdb(content []byte) error {
	if err := conn.w.WriteRdb(content); err != nil {
		return err
	}
	return conn.w.Flush()
}

func (conn *Conn) WriteInt(v int) error {
	if err := conn.w.WriteInt(v); err != nil {
		return err
	}
	return conn.w.Flush()
}

func (conn *Conn) WriteErrorInvalidCmd() error {
	if err := conn.w.WriteError([]byte("Invalid Command")); err != nil {
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
	fmt.Printf("closing connection: %v->%v\n", conn.netConn.LocalAddr(), conn.netConn.RemoteAddr())
	return conn.netConn.Close()
}
