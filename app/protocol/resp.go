package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// Frame is a single RESP message node that can encode itself to a writer.
type Frame interface {
	WriteTo(w *bufio.Writer) error
}

// SimpleString represents a RESP Simple String: +OK\r\n
type SimpleString struct {
	Value string
}

func (s SimpleString) WriteTo(w *bufio.Writer) error {
	if _, err := w.WriteString("+" + s.Value + "\r\n"); err != nil {
		return err
	}
	return nil
}

// Error represents a RESP Error: -ERR message\r\n
type Error struct {
	Message string
}

func (e Error) WriteTo(w *bufio.Writer) error {
	if _, err := w.WriteString("-ERR " + e.Message + "\r\n"); err != nil {
		return err
	}
	return nil
}

// BulkString represents a RESP Bulk String: $<len>\r\n<data>\r\n
type BulkString struct {
	Bytes []byte
}

func (b BulkString) WriteTo(w *bufio.Writer) error {
	if _, err := fmt.Fprintf(w, "$%d\r\n", len(b.Bytes)); err != nil {
		return err
	}
	if _, err := w.Write(b.Bytes); err != nil {
		return err
	}
	if _, err := w.WriteString("\r\n"); err != nil {
		return err
	}
	return nil
}

// BulkNullString represents a RESP Null Bulk String: $-1\r\n
type BulkNullString struct{}

func (BulkNullString) WriteTo(w *bufio.Writer) error {
	_, err := w.WriteString("$-1\r\n")
	return err
}

// Array represents a RESP Array. When Null is true, it encodes as *-1\r\n
type Array struct {
	Elems []Frame
	Null  bool
}

func (a Array) WriteTo(w *bufio.Writer) error {
	if a.Null {
		_, err := w.WriteString("*-1\r\n")
		return err
	}
	if _, err := fmt.Fprintf(w, "*%d\r\n", len(a.Elems)); err != nil {
		return err
	}
	for _, el := range a.Elems {
		if err := el.WriteTo(w); err != nil {
			return err
		}
	}
	return nil
}

// WriteFrame encodes f to w and flushes.
func WriteFrame(w *bufio.Writer, f Frame) error {
	if err := f.WriteTo(w); err != nil {
		return err
	}
	return w.Flush()
}

// ReadFrame parses a single RESP frame from r.
func ReadFrame(r *bufio.Reader) (Frame, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case '+': // Simple String
		line, err := readCRLFLine(r)
		if err != nil {
			return nil, err
		}
		return SimpleString{Value: line}, nil

	case '-': // Error
		line, err := readCRLFLine(r)
		if err != nil {
			return nil, err
		}
		return Error{Message: line}, nil

	case '$': // Bulk String
		nStr, err := readCRLFLine(r)
		if err != nil {
			return nil, err
		}
		n, err := strconv.Atoi(nStr)
		if err != nil {
			return nil, fmt.Errorf("invalid bulk length")
		}
		if n == -1 {
			return BulkNullString{}, nil
		}
		if n < 0 {
			return nil, fmt.Errorf("invalid bulk length")
		}
		// read n bytes + CRLF
		buf := make([]byte, n+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		if buf[n] != '\r' || buf[n+1] != '\n' {
			return nil, fmt.Errorf("bulk data missing CRLF")
		}
		return BulkString{Bytes: buf[:n]}, nil

	case '*': // Array
		nStr, err := readCRLFLine(r)
		if err != nil {
			return nil, err
		}
		n, err := strconv.Atoi(nStr)
		if err != nil {
			return nil, fmt.Errorf("invalid array length")
		}
		if n == -1 {
			return Array{Null: true}, nil
		}
		if n < 0 {
			return nil, fmt.Errorf("invalid array length")
		}
		elems := make([]Frame, 0, n)
		for i := 0; i < n; i++ {
			el, err := ReadFrame(r)
			if err != nil {
				return nil, err
			}
			elems = append(elems, el)
		}
		return Array{Elems: elems}, nil

	default:
		return nil, fmt.Errorf("unknown RESP type: %q", b)
	}
}

// readCRLFLine reads a line terminated with CRLF and returns the line without CRLF.
func readCRLFLine(r *bufio.Reader) (string, error) {
	s, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	// s must end with \r\n
	if len(s) < 2 || s[len(s)-2] != '\r' || s[len(s)-1] != '\n' {
		return "", fmt.Errorf("malformed line")
	}
	return s[:len(s)-2], nil
}
