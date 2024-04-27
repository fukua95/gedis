package server

import (
	"fmt"
	"io"
	"net"
	"strconv"

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
	port    string
	addr    string
	store   *storage.Store

	role role

	masterReplID     string
	masterReplOffset int
	slavePort        string
	slaveCapa        string

	masterAddr string
}

func NewServer(conf *Config) *Server {
	s := &Server{
		network:    conf.network,
		port:       conf.port,
		addr:       conf.addr,
		store:      storage.NewStore(),
		masterAddr: conf.masterAddr,
	}
	if conf.isSlave {
		s.role = roleSlave
		s.slavePort = s.port
	} else {
		s.role = roleMaster
		s.masterReplID = util.RandomAlphanumericString(40)
		s.masterReplOffset = 0
	}
	return s
}

func (s *Server) ListenAndServe() error {
	if s.role == roleSlave {
		go s.replicate()
	}

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
		conn := NewConn(c)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn *Conn) {
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
			err = s.set(conn, cmd)
		case resp.CmdGet:
			err = s.get(conn, cmd)
		case resp.CmdInfo:
			err = s.info(conn, cmd)
		case resp.CmdRepl:
			err = s.replconf(conn, cmd)
		case resp.CmdPsync:
			err = s.psync(conn, cmd)
		}
		if err != nil {
			fmt.Println("Error handle command: ", err.Error())
			return
		}
	}
}

func (s *Server) set(conn *Conn, cmd Command) error {
	args := cmd.Args()
	if len(args) < 3 {
		return conn.WriteErrorInvalidCmd()
	}
	if px, ok := cmd.SearchOption(resp.OptionSetEx); ok {
		s.store.PutEx(args[1], args[2], px)
	} else {
		s.store.Put(args[1], args[2])
	}
	return conn.WriteStatusOK()
}

func (s *Server) get(conn *Conn, cmd Command) error {
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

func (s *Server) info(conn *Conn, _ Command) error {
	info := fmt.Sprintf("role:%s", s.role)
	if s.role == roleMaster {
		info = fmt.Sprintf("%s\nmaster_replid:%s\nmaster_repl_offset:%s",
			info, s.masterReplID, util.Itoa(s.masterReplOffset))
	}
	return conn.WriteString(info)
}

func (s *Server) replconf(conn *Conn, cmd Command) error {
	if port, has := cmd.SearchOption(resp.OptionReplLPort); has {
		s.slavePort = string(port[0])
	}
	if capa, has := cmd.SearchOption(resp.OptionReplCapa); has {
		s.slaveCapa = string(capa[0])
	}
	return conn.WriteStatusOK()
}

func (s *Server) psync(conn *Conn, _ Command) error {
	status := fmt.Sprintf("FULLRESYNC %s %s", s.masterReplID, strconv.Itoa(s.masterReplOffset))
	return conn.WriteStatus([]byte(status))
}

func (s *Server) replicate() {
	c, err := net.Dial(s.network, s.masterAddr)
	if err != nil {
		fmt.Println("slave connect to master failed: ", err.Error())
		return
	}
	defer c.Close()

	conn := NewConn(c)
	if err = s.handshake(conn); err != nil {
		fmt.Println("slave handshake with master failed: ", err.Error())
		return
	}

}

func (s *Server) handshake(conn *Conn) error {
	cmd := &command{args: [][]byte{[]byte(resp.CmdPing)}}
	if err := conn.WriteCommand(cmd); err != nil {
		return err
	}

	cmd = &command{
		args: [][]byte{
			[]byte(resp.CmdRepl),
			[]byte(resp.OptionReplLPort),
			[]byte(s.port),
		},
	}
	if err := conn.WriteCommand(cmd); err != nil {
		return err
	}

	cmd = &command{
		args: [][]byte{
			[]byte(resp.CmdRepl),
			[]byte(resp.OptionReplCapa),
			[]byte("psync2"),
		},
	}
	if err := conn.WriteCommand(cmd); err != nil {
		return err
	}

	cmd = &command{
		args: [][]byte{
			[]byte(resp.CmdPsync),
			[]byte("?"),
			[]byte("-1"),
		},
	}
	return conn.WriteCommand(cmd)
}
