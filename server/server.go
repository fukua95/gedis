package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/rdb"
	"github.com/codecrafters-io/redis-starter-go/resp"
	"github.com/codecrafters-io/redis-starter-go/storage"
	"github.com/codecrafters-io/redis-starter-go/util"
)

type role string

const (
	roleMaster  role = "master"
	roleReplica role = "slave"
)

type Server struct {
	network    string
	port       string
	addr       string
	store      *storage.Store
	dir        string
	dbfilename string

	role       role
	replID     string
	replOffset int

	// sync write cmd to store and propagate to replicas.
	mu sync.Mutex

	// for master
	replicas *storage.SyncSlice[*Conn]
	propCh   chan Command

	// for replica
	masterAddr string
}

func NewServer(conf *Config) *Server {
	s := &Server{
		network:    conf.network,
		port:       conf.port,
		addr:       conf.addr,
		store:      storage.NewStore(),
		dir:        conf.dir,
		dbfilename: conf.dbfilename,
		role:       conf.role,
	}

	s.loadRdb()

	if s.role == roleMaster {
		s.replID = util.RandomAlphanumericString(40)
		s.replOffset = 0
		s.propCh = make(chan Command, 10)
		s.replicas = new(storage.SyncSlice[*Conn])
		go s.asMaster()
	} else {
		s.masterAddr = conf.masterAddr
		go s.asReplica()
	}
	return s
}

func (s *Server) loadRdb() {
	if s.dir == "" || s.dbfilename == "" {
		return
	}

	db := fmt.Sprintf("%s/%s", s.dir, s.dbfilename)
	f, err := os.Open(db)
	if err != nil {
		fmt.Printf("read file %s error %s\n", db, err.Error())
		return
	}

	// kvCh := make(chan rdb.Entry, 100)
	kvCh := make(chan rdb.Entry, 100)
	rdb := rdb.NewRdb(f)
	go rdb.Read(kvCh)

	for kv := range kvCh {
		s.store.Put([]byte(kv.K), []byte(kv.V), int64(kv.Ex))
	}

	f.Close()
	fmt.Println("server successfully loaded rdb")
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
	isReplica := false
	defer func() {
		if !isReplica {
			conn.Close()
		}
	}()

	for {
		cmd, err := conn.ReadCommand()
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Printf("role=%s Error reading from conn: %q\n", s.role, err.Error())
			return
		}

		switch cmd.Name() {
		case resp.CmdEcho:
			err = conn.WriteString(string(cmd.At(1)))
		case resp.CmdPing:
			err = conn.WriteString("PONG")
		case resp.CmdSet:
			err = s.set(conn, cmd)
			if err != nil {
				conn.WriteErrorInvalidCmd()
			}
			conn.WriteStatusOK()
		case resp.CmdGet:
			err = s.get(conn, cmd)
		case resp.CmdInfo:
			err = s.info(conn, cmd)
		case resp.CmdReplConf:
			err = s.replconf(conn, cmd)
		case resp.CmdPsync:
			err = s.psync(conn, cmd)
			isReplica = true
		case resp.CmdWait:
			err = s.wait(conn, cmd)
		case resp.CmdConfig:
			err = s.config(conn, cmd)
		case resp.CmdKeys:
			err = s.keys(conn, cmd)
		}
		if err != nil {
			fmt.Println("Error handle command: ", err.Error())
			return
		}
		if isReplica {
			break
		}
	}
}

func (s *Server) set(_ *Conn, cmd Command) error {
	args := cmd.Args()
	if len(args) < 3 {
		return errors.New("invalid command")
	}
	px, hasPx := cmd.SearchOption(resp.OptionSetEx)
	ex := 0
	var err error
	if hasPx {
		if ex, err = util.Atoi(px); err != nil {
			return err
		}
		ex += int(time.Now().UnixMilli())
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.store.Put(args[1], args[2], int64(ex))
	s.propagate(cmd)

	return nil
}

func (s *Server) get(conn *Conn, cmd Command) error {
	args := cmd.Args()
	if len(args) < 2 {
		return conn.WriteErrorInvalidCmd()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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
	status := fmt.Sprintf("%s %s %s", resp.ReplyFullResync, s.replID, strconv.Itoa(int(s.replOffset)))
	if err := conn.WriteStatus(status); err != nil {
		return err
	}
	// send a rdb file.
	if err := conn.WriteRdb([]byte(rdb.EmptyRdb())); err != nil {
		return err
	}
	fmt.Println("master finishes sending rdb file")

	s.replicas.Append(conn)
	return nil
}

// `wait` waits until:
// - the expected number of replicas complete sync with master,
// - or timeout expires.
// `wait` should return the number of replicas that sync with master, even if the timeout expires.
func (s *Server) wait(conn *Conn, cmd Command) error {
	if len(cmd.Args()) != 3 {
		return resp.ErrInvalidCommand
	}
	threshold, _ := util.Atoi(cmd.At(1))
	timeoutMS, _ := util.Atoi(cmd.At(2))

	if s.replOffset == 0 {
		return conn.WriteInt(s.replicas.Len())
	}

	if threshold <= 0 || timeoutMS <= 0 {
		return conn.WriteInt(0)
	}

	replicas := s.replicas.Clone()
	isSync := make(chan int, len(replicas)+1)
	syncCount := 0
	getAckCmd := &command{args: [][]byte{[]byte(resp.CmdReplConf), []byte(resp.OptionGetAck), []byte("*")}}
	timeout := time.Now().Add(time.Duration(timeoutMS) * time.Millisecond)

	for _, replica := range replicas {
		go func(conn *Conn, isSync chan<- int, offset int, cmd *command, timeout time.Time) {
			isSync <- func() int {
				defer conn.ResetReadDeadline()

				err := conn.WriteCommand(cmd)
				if err != nil {
					fmt.Println("master send getack error: ", err.Error())
					return 0
				}
				conn.SetReadDeadline(timeout)
				reply, err := conn.ReadSliceReply()
				if err != nil {
					fmt.Println("master getack reply error: ", err.Error())
					return 0
				}

				if len(reply) != 3 || string(reply[0]) != resp.CmdReplConf || string(reply[1]) != resp.OptionAck {
					fmt.Println("master getack reply error: invalid reply")
					return 0
				}

				replicaOffset, _ := util.Atoi(reply[2])
				if replicaOffset < offset {
					fmt.Printf("master offset=%v, replica offset=%v\n", offset, replicaOffset)
					return 0
				}

				fmt.Println("master getack from one replica successfully, offset=", offset)
				return 1
			}()
		}(replica, isSync, s.replOffset, getAckCmd, timeout)
	}

	for i := 0; i < len(replicas); i++ {
		syncCount += <-isSync
	}

	fmt.Println("master getack count=", syncCount)

	s.replOffset += getAckCmd.RespLen()

	return conn.WriteInt(syncCount)
}

func (s *Server) config(conn *Conn, cmd Command) error {
	var reply [][]byte
	switch string(cmd.At(2)) {
	case resp.OptionDir:
		reply = [][]byte{[]byte(resp.OptionDir), []byte(s.dir)}
	case resp.OptionDBFile:
		reply = [][]byte{[]byte(resp.OptionDBFile), []byte(s.dbfilename)}
	}
	return conn.WriteSlice(reply)
}

func (s *Server) keys(conn *Conn, _ Command) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := s.store.Scan()
	strKeys := make([][]byte, len(keys))
	for i, k := range keys {
		strKeys[i] = []byte(k)
	}
	return conn.WriteSlice(strKeys)
}

func (s *Server) propagate(cmd Command) {
	if s.role == roleMaster {
		s.propCh <- cmd
		s.replOffset += cmd.RespLen()
	}
}

func (s *Server) asMaster() {
	for cmd := range s.propCh {
		replicas := s.replicas.Clone()
		for _, replica := range replicas {
			replica.WriteCommand(cmd)
		}
	}
}

func (s *Server) asReplica() {
	c, err := net.Dial(s.network, s.masterAddr)
	if err != nil {
		fmt.Println("replica connect to master error: ", err.Error())
		return
	}

	conn := NewConn(c)
	defer conn.Close()

	if err = s.handshake(conn); err != nil {
		fmt.Println("replica handshake with master error: ", err.Error())
		return
	}

	if err = s.requestFullResync(conn); err != nil {
		fmt.Println("replica full resynchronization error: ", err.Error())
		return
	}

	for {
		cmd, err := conn.ReadCommand()
		if err != nil {
			fmt.Println("Error reading from master: ", err.Error())
			break
		}
		// master -> replica, replica 只回复 REPLCONF, 其余 cmd 不回复.
		switch cmd.Name() {
		case resp.CmdSet:
			s.set(conn, cmd)
		case resp.CmdReplConf:
			if len(cmd.Args()) != 3 || string(cmd.At(1)) != resp.OptionGetAck {
				fmt.Println("Error reading from master: invalid REPLCONF command")
				conn.WriteErrorInvalidCmd()
				return
			}
			fmt.Println("replica receives GETACK command from master")
			reply := [][]byte{[]byte(resp.CmdReplConf), []byte("ACK"), util.Itoa(s.replOffset)}
			if err := conn.WriteSlice(reply); err != nil {
				fmt.Println("replica reply GETACK error: ", err.Error())
			} else {
				fmt.Println("replica reply GETACK successfully")
			}
		}
		s.replOffset += cmd.RespLen()
	}
}

func (s *Server) handshake(conn *Conn) error {
	cmd := &command{args: [][]byte{[]byte(resp.CmdPing)}}
	if err := s.WriteCmdAndCheckReply(conn, cmd, "pong"); err != nil {
		return err
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
	if err := conn.WriteCommand(cmd); err != nil {
		return err
	}
	replyStr, err := conn.ReadStatusReply()
	if err != nil {
		return err
	}
	reply := strings.Split(replyStr, " ")
	if len(reply) != 3 || reply[0] != resp.ReplyFullResync {
		return resp.ErrInvalidReply
	}
	s.replID = reply[1]
	s.replOffset, _ = strconv.Atoi(reply[2])

	// read the rdb file from the master, and apply the rdb file.
	_, err = conn.ReadRdb()
	fmt.Println("replica finishes receiving rdb file")
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
	if !strings.EqualFold(v, reply) {
		return resp.ErrInvalidReply
	}
	return nil
}
