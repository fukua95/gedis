package server

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

type Server struct {
	network string
	address string
	kv      sync.Map
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) ListenAndServe(network string, address string) error {
	s.network = network
	s.address = address
	l, err := net.Listen(network, address)
	if err != nil {
		fmt.Println("listen error: ", err.Error())
		return err
	}
	defer l.Close()

	fmt.Println("Server start to accept requests")

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("accept error: ", err.Error())
			return err
		}
		conn := resp.NewConn(c)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn *resp.Conn) {
	defer conn.Close()

	for {
		cmd, err := conn.ReadCommand()
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Println("Error reading from conn: ", err.Error())
			return
		}
		switch cmd.Name() {
		case resp.CmdEcho:
			args := cmd.Args()
			args = args[1:]
			if len(args) == 1 {
				err = conn.WriteString(string(args[0]))
			} else {
				err = conn.WriteArray(args[1:])
			}
		case resp.CmdPing:
			err = conn.WriteString("PONG")
		case resp.CmdSet:
			err = s.handleCmdSet(conn, cmd)
		case resp.CmdGet:
			err = s.handleCmdGet(conn, cmd)
		}
		if err != nil {
			fmt.Println("Error handle command: ", err.Error())
			return
		}
	}
}

func (s *Server) handleCmdSet(conn *resp.Conn, cmd resp.Command) error {
	args := cmd.Args()
	if len(args) < 3 {
		return conn.WriteErrorInvalidCmd()
	}
	// golang's func, map, slice 不支持 hash, 所以不能作为 map/sync.map 的 key.
	// arg 是 []byte, 先转化为 string.
	s.kv.Store(string(args[1]), string(args[2]))
	return conn.WriteStatusOK()
}

func (s *Server) handleCmdGet(conn *resp.Conn, cmd resp.Command) error {
	args := cmd.Args()
	if len(args) < 2 {
		return conn.WriteErrorInvalidCmd()
	}
	val, ok := s.kv.Load(string(args[1]))
	if !ok {
		return conn.WriteNilBulkString()
	}
	return conn.WriteString(val.(string))
}
