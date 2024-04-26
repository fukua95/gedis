package server

import (
	"fmt"
	"io"
	"net"

	"github.com/codecrafters-io/redis-starter-go/resp"
	"github.com/codecrafters-io/redis-starter-go/storage"
	"github.com/codecrafters-io/redis-starter-go/util"
)

type Server struct {
	network string
	address string
	store   *storage.Store
	repl    *replication
}

func NewServer(conf *Config) *Server {
	return &Server{
		network: conf.network,
		address: conf.addr,
		repl:    newReplication(conf.isSlave, conf.masterAddr),
		store:   storage.NewStore(),
	}
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen(s.network, s.address)
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
	info := [][]byte{
		[]byte(fmt.Sprintf("role:%s", s.repl.role)),
	}
	if s.repl.role == roleMaster {
		info = append(info, []byte(fmt.Sprintf("master_replid:%s", s.repl.masterReplID)))
		info = append(info, []byte(fmt.Sprintf("master_repl_offset:%s", util.Itoa(s.repl.masterReplOffset))))
	}
	return conn.WriteArray(info)
}
