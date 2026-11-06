// Package sse provides bounded event framing for server-sent event streams.
package sse

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

const MaxEventSize = 1 << 20

var ErrEventTooLarge = errors.New("sse: event exceeds 1 MiB")

type Reader struct {
	r   *bufio.Reader
	max int
}

func NewReader(r io.Reader) *Reader { return &Reader{r: bufio.NewReader(r), max: MaxEventSize} }

// ReadEvent returns one event without its terminating blank line. It accepts LF and CRLF.
// EOF after a complete event is reported on the next call, as required by io.Reader conventions.
func (r *Reader) ReadEvent() ([]byte, error) {
	var event []byte
	for {
		line, err := r.r.ReadBytes('\n')
		if len(line) > 0 {
			line = bytes.TrimSuffix(line, []byte{'\n'})
			line = bytes.TrimSuffix(line, []byte{'\r'})
			if len(line) == 0 {
				return append([]byte(nil), event...), nil
			}
			event = append(event, line...)
			if len(event) > r.max {
				return nil, ErrEventTooLarge
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) && len(event) > 0 {
				return append([]byte(nil), event...), io.EOF
			}
			return nil, err
		}
		event = append(event, '\n')
		if len(event) > r.max {
			return nil, ErrEventTooLarge
		}
	}
}

type Writer struct{ w io.Writer }

func NewWriter(w io.Writer) *Writer { return &Writer{w: w} }
func (w *Writer) WriteEvent(event []byte) error {
	_, err := w.w.Write(append(append([]byte(nil), event...), '\n', '\n'))
	return err
}
