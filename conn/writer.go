package conn

import (
	"bufio"
	"fmt"
	"io"

	"github.com/codecrafters-io/redis-starter-go/util"
)

type Writer struct {
	w bufio.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: *bufio.NewWriter(w),
	}
}

func (w *Writer) WriteStatus(b []byte) error {
	line := fmt.Sprintf("%c%s\r\n", RespStatus, b)
	_, err := w.w.Write([]byte(line))
	return err
}

func (w *Writer) WriteBytes(b []byte) error {
	line := fmt.Sprintf("%c%s\r\n%s\r\n", RespString, util.Itoa(len(b)), b)
	_, err := w.w.Write([]byte(line))
	return err
}

func (w *Writer) WriteArray(arr [][]byte) error {
	line := fmt.Sprintf("%c%s\r\n", RespArray, util.Itoa(len(arr)))
	_, err := w.w.Write([]byte(line))
	if err != nil {
		return err
	}
	for _, b := range arr {
		err = w.WriteBytes(b)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
func (w *Writer) writeEnd() error {
	err := w.w.WriteByte('\r')
	if err != nil {
		return err
	}
	return w.w.WriteByte('\n')
}
*/

func (w *Writer) Flush() error {
	return w.w.Flush()
}
