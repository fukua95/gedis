package server

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/resp"
	"github.com/codecrafters-io/redis-starter-go/storage"
	"github.com/codecrafters-io/redis-starter-go/util"
)

type role string

const (
	roleMaster  role = "master"
	roleReplica role = "replica"
)

type Server struct {
	network string
	port    string
	addr    string
	store   *storage.Store

	role role

	// for master
	replID     string
	replOffset int

	// for replica
	masterAddr string
}

func NewServer(conf *Config) *Server {
	s := &Server{
		network: conf.network,
		port:    conf.port,
		addr:    conf.addr,
		store:   storage.NewStore(),
		role:    conf.role,
	}
	if s.role == roleMaster {
		s.replID = util.RandomAlphanumericString(40)
		s.replOffset = 0
	}
	if s.role == roleReplica {
		s.masterAddr = conf.masterAddr
		go s.asReplica()
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
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept error: ", err.Error())
			return err
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(c net.Conn) {
	conn := NewConn(c)
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
			err = conn.WriteString(string(cmd.At(1)))
		case resp.CmdPing:
			err = conn.WriteString("PONG")
		case resp.CmdSet:
			err = s.set(conn, cmd)
		case resp.CmdGet:
			err = s.get(conn, cmd)
		case resp.CmdInfo:
			err = s.info(conn, cmd)
		case resp.CmdReplConf:
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
			info, s.replID, util.Itoa(s.replOffset))
	}
	return conn.WriteString(info)
}

func (s *Server) replconf(conn *Conn, _ Command) error {
	return conn.WriteStatusOK()
}

func (s *Server) psync(conn *Conn, cmd Command) error {
	if len(cmd.Args()) != 3 {
		panic("psync cmd is invalid")
	}
	// replicaReplID := cmd.At(1)
	// offset := cmd.At(2)
	// psync repl_id, offset 表示 replica 希望 master(repl_id = repl_id) 从 offset 开始继续同步.
	// repl_id = ? 表示 replica 第一次连接到这个 master, 不知道 master's repl_id.
	// repl_id != ? 时, 检查 master's repl_id = repl_id.
	// offset = -1, 表示从头开始同步: 发送 rdb file + 后续同步.
	// 这里是 `psync ? -1`, 所以先忽略相关逻辑.
	status := fmt.Sprintf("%s %s %s", resp.ReplyFullResync, s.replID, strconv.Itoa(s.replOffset))
	if err := conn.WriteStatus([]byte(status)); err != nil {
		return err
	}
	// send a rdb file.
	if err := conn.WriteRdb([]byte(EmptyRdb())); err != nil {
		return err
	}
	return nil
}

func (s *Server) asReplica() {
	c, err := net.Dial(s.network, s.masterAddr)
	if err != nil {
		fmt.Println("replica connect to master error: ", err.Error())
		return
	}
	defer c.Close()

	conn := NewConn(c)

	if err = s.handshake(conn); err != nil {
		fmt.Println("replica handshake with master error: ", err.Error())
		return
	}

	if err = s.requestFullResync(conn); err != nil {
		fmt.Println("replica full resynchronization error: ", err.Error())
		return
	}
}

func (s *Server) handshake(conn *Conn) error {
	cmd := &command{args: [][]byte{[]byte(resp.CmdPing)}}
	if err := s.WriteCmdAndCheckReply(conn, cmd, "pong"); err != nil {
		return nil
	}

	cmd = &command{
		args: [][]byte{
			[]byte(resp.CmdReplConf),
			[]byte(resp.OptionReplLPort),
			[]byte(s.port),
		},
	}
	if err := s.WriteCmdAndCheckReply(conn, cmd, "ok"); err != nil {
		return err
	}

	cmd = &command{
		args: [][]byte{
			[]byte(resp.CmdReplConf),
			[]byte(resp.OptionReplCapa),
			[]byte("psync2"),
		},
	}
	if err := s.WriteCmdAndCheckReply(conn, cmd, "ok"); err != nil {
		return err
	}
	return nil
}

func (s *Server) requestFullResync(conn *Conn) error {
	// replica sens a `PSYNC ? -1` to tell the master that it doesn't have any data,
	// and needs to be full resynchronized.
	cmd := &command{args: [][]byte{[]byte(resp.CmdPsync), []byte("?"), []byte("-1")}}
	if err := s.WriteCmdAndCheckReply(conn, cmd, resp.ReplyFullResync); err != nil {
		return err
	}
	_, err := conn.ReadBytesReply()
	return err
}

func (s *Server) WriteCmdAndCheckReply(conn *Conn, cmd Command, reply string) error {
	err := conn.WriteCommand(cmd)
	if err != nil {
		return err
	}
	v, err := conn.ReadStatusReply()
	if err != nil {
		return err
	}
	if strings.EqualFold(v, reply) {
		return resp.ErrInvalidReply
	}
	return nil
}
