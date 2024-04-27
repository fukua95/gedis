package server

import (
	"fmt"
	"io"
	"net"

	"github.com/codecrafters-io/redis-starter-go/resp"
	"github.com/codecrafters-io/redis-starter-go/storage"
	"github.com/codecrafters-io/redis-starter-go/util"
)

type role string

const (
	roleMaster role = "master"
	roleSlave  role = "slave"
)

type Server struct {
	network string
	addr    string
	store   *storage.Store

	role             role
	masterReplID     string
	masterReplOffset int
	masterAddr       string
}

func NewServer(conf *Config) *Server {
	s := &Server{
		network:    conf.network,
		addr:       conf.addr,
		store:      storage.NewStore(),
		masterAddr: conf.masterAddr,
	}
	if conf.isSlave {
		s.role = roleSlave
	} else {
		s.role = roleMaster
		s.masterReplID = util.RandomAlphanumericString(40)
		s.masterReplOffset = 0
	}
	return s
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen(s.network, s.addr)
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
				err = conn.WriteSlice(args[1:])
			}
		case resp.CmdPing:
			err = conn.WriteString("PONG")
		case resp.CmdSet:
			err = s.handleCmdSet(conn, cmd)
		case resp.CmdGet:
			err = s.handleCmdGet(conn, cmd)
		case resp.CmdInfo:
			err = s.handleCmdInfo(conn, cmd)
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
	if px, ok := cmd.OptionSetEx(); ok {
		s.store.PutEx(args[1], args[2], px)
	} else {
		s.store.Put(args[1], args[2])
	}
	return conn.WriteStatusOK()
}

func (s *Server) handleCmdGet(conn *resp.Conn, cmd resp.Command) error {
	args := cmd.Args()
	if len(args) < 2 {
		return conn.WriteErrorInvalidCmd()
	}
	val, ok := s.store.Get(args[1])
	if !ok {
		return conn.WriteNilBulkString()
	}
	return conn.WriteString(string(val))
}

func (s *Server) handleCmdInfo(conn *resp.Conn, _ resp.Command) error {
	info := fmt.Sprintf("role:%s", s.role)
	if s.role == roleMaster {
		info = fmt.Sprintf("%s\nmaster_replid:%s\nmaster_repl_offset:%s",
			info, s.masterReplID, util.Itoa(s.masterReplOffset))
	}
	return conn.WriteString(info)
}
