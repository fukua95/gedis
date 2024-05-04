package resp

import (
	"bufio"
	"io"
)

type Writer struct {
	w bufio.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: *bufio.NewWriter(w),
	}
}

func (w *Writer) WriteStatus(b string) error {
	_, err := w.w.Write(Status(b))
	return err
}

func (w *Writer) WriteError(e string) error {
	_, err := w.w.Write(Error(e))
	return err
}

func (w *Writer) WriteNilBulkString() error {
	_, err := w.w.Write(NilString())
	return err
}

func (w *Writer) WriteBytes(b []byte) error {
	_, err := w.w.Write(String(string(b)))
	return err
}

func (w *Writer) WriteSlice(arr [][]byte) error {
	_, err := w.w.Write(ArrayHeader(len(arr)))
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

func (w *Writer) WriteRdb(content []byte) error {
	_, err := w.w.Write(RdbContent(content))
	return err
}

func (w *Writer) WriteInt(v int) error {
	_, err := w.w.Write(Integer(v))
	return err
}

func (w *Writer) WriteRawBytes(b []byte) error {
	_, err := w.w.Write(b)
	return err
}

func (w *Writer) Flush() error {
	return w.w.Flush()
}
