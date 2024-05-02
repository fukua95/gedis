package rdb

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/codecrafters-io/redis-starter-go/util"
)

var (
	// the contents of an empty rdb file in base64
	emptyRdbBase64 = "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
)

func EmptyRdb() []byte {
	rdb, err := base64.StdEncoding.DecodeString(emptyRdbBase64)
	if err != nil {
		log.Fatalln(err)
	}
	return rdb
}

const (
	EOF          uint8 = 0xFF
	SELECTDB     uint8 = 0xFE
	EXPIRETIME   uint8 = 0xFD
	EXPIRETIMEMS uint8 = 0xFC
	RESIZEDB     uint8 = 0xFB
	AUX          uint8 = 0xFA
)

const (
	String             uint8 = 0
	List               uint8 = 1
	Set                uint8 = 2
	SortedSet          uint8 = 3
	Hash               uint8 = 4
	Zipmap             uint8 = 9
	Ziplist            uint8 = 10
	IntSet             uint8 = 11
	SortedSetInZiplist uint8 = 12
	HashmapInZiplist   uint8 = 13
	ListInQuicklist    uint8 = 14
)

var ErrDBEOF = errors.New("DB EOF")

type Entry struct {
	K  string
	V  string
	Ex int64
}

type Rdb struct {
	f *os.File
}

func NewRdb(f *os.File) *Rdb {
	return &Rdb{f: f}
}

func (r *Rdb) readBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	l, err := r.f.Read(b)
	if err == io.EOF {
		fmt.Println("read rdb EOF")
		return nil, err
	}
	if err != nil {
		fmt.Println("read rdb error: ", err.Error())
		return nil, err
	}
	fmt.Printf("read %v str=%q, bytes=%v\n", l, string(b), b)
	if l != n {
		fmt.Printf("read byte size error, wanted=%v, has=%v\n", n, l)
		return nil, errors.New("read size error")
	}
	return b, nil
}

func (r *Rdb) readByte() (byte, error) {
	b, err := r.readBytes(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (r *Rdb) readHeader() (string, int, error) {
	b, err := r.readBytes(9)
	if err != nil {
		return "", 0, err
	}
	version, _ := util.Atoi(b[5:9])
	return string(b[0:5]), version, nil
}

// rdb file format: https://rdb.fnordig.de/file_format.html
func (r *Rdb) Read(kvCh chan<- Entry) {
	defer func() {
		close(kvCh)
	}()

	magic, version, err := r.readHeader()
	if err != nil {
		return
	}
	fmt.Printf("rdb magic number=%s, version=%v\n", magic, version)

	for {
		b, err := r.readByte()
		if err != nil {
			return
		}
		switch b {
		case AUX:
			for i := 0; i < 2; i++ {
				if _, err := r.readString(); err != nil {
					return
				}
			}
		case SELECTDB:
			if err := r.readData(kvCh); err != nil {
				return
			}
		case EOF:
			return
		default:
			panic(fmt.Sprintf("read impossible byte=%v", b))
		}
	}
}

func (r *Rdb) readData(kvCh chan<- Entry) error {
	dbNum, err := r.readLen()
	if err != nil {
		return err
	}
	b, err := r.readByte()
	if err != nil {
		return err
	}

	if b != RESIZEDB {
		panic("missing FB op codes")
	}
	for i := 0; i < 2; i++ {
		_, err = r.readLen()
		if err != nil {
			return err
		}
	}

	fmt.Println("reading data from db number = ", dbNum)
	for {
		entry, err := r.readEntry()
		if err == ErrDBEOF {
			return nil
		}
		if err != nil {
			return err
		}
		kvCh <- entry
	}
}

func (r *Rdb) readEntry() (Entry, error) {
	t, err := r.readByte()
	if err != nil {
		return Entry{}, nil
	}

	switch t {
	case EOF, SELECTDB:
		r.f.Seek(-1, 1)
		return Entry{}, ErrDBEOF
	case EXPIRETIME:
		ex, err := r.readBytes(4)
		if err != nil {
			return Entry{}, err
		}
		e, _ := r.readEntry()
		e.Ex = int64(binary.LittleEndian.Uint32(ex))
		return e, nil
	case EXPIRETIMEMS:
		ex, err := r.readBytes(8)
		if err != nil {
			return Entry{}, err
		}
		e, _ := r.readEntry()
		e.Ex = int64(binary.LittleEndian.Uint64(ex))
		return e, nil
	case String:
		k, err := r.readString()
		if err != nil {
			return Entry{}, err
		}
		v, err := r.readString()
		if err != nil {
			return Entry{}, err
		}
		e := Entry{K: k, V: v, Ex: 0}
		return e, nil
	}
	// ignore value types: set, map, &c.
	panic("rdb file has other value type")
}

func (r *Rdb) readString() (string, error) {
	l, err := r.readLen()
	if err != nil {
		return "", err
	}
	b, err := r.readBytes(l)
	return string(b), err
}

func (r *Rdb) readLen() (int, error) {
	b, err := r.readByte()
	if err != nil {
		return 0, err
	}
	t := uint8(b) >> 6
	switch t {
	case 0:
		return int(uint8(b)), nil
	case 3:
		tl := uint8(b) - (1 << 7) - (1 << 6)
		switch tl {
		case 0:
			return 1, nil
		case 1:
			return 2, nil
		case 2:
			return 4, nil
		default:
			panic(fmt.Sprintf("rdb file has other len encoding cases, t=%v, b=%v", t, uint8(b)))
		}
	default:
		panic(fmt.Sprintf("rdb file has other len encoding cases, t=%v, b=%v", t, uint8(b)))
	}
}
