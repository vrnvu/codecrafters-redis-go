package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// Frame is a single RESP message node that can encode itself to a writer.
type Frame interface {
	Write(w *bufio.Writer) error
}

// SimpleString represents a RESP Simple String: +OK\r\n
type SimpleString struct {
	Value string
}

func (s SimpleString) Write(w *bufio.Writer) error {
	n, err := w.WriteString("+" + s.Value + "\r\n")
	if err != nil {
		return err
	}
	if n != len(s.Value)+3 {
		return fmt.Errorf("expected to write %d bytes, wrote %d", len(s.Value)+3, n)
	}

	return w.Flush()
}

// Error represents a RESP Error: -ERR message\r\n
type Error struct {
	Message string
}

func (e Error) Write(w *bufio.Writer) error {
	n, err := w.WriteString("-ERR " + e.Message + "\r\n")
	if err != nil {
		return err
	}
	if n != len(e.Message)+7 {
		return fmt.Errorf("expected to write %d bytes, wrote %d", len(e.Message)+7, n)
	}

	return w.Flush()
}

// BulkString represents a RESP Bulk String: $<len>\r\n<data>\r\n
type BulkString struct {
	Bytes []byte
}

func (b BulkString) Write(w *bufio.Writer) error {
	n, err := fmt.Fprintf(w, "$%d\r\n", len(b.Bytes))
	if err != nil {
		return err
	}
	// The header is "$<digits>\r\n" - we need to calculate actual digit length
	expectedHeaderLen := len(fmt.Sprintf("%d", len(b.Bytes))) + 3 // $ + digits + \r\n
	if n != expectedHeaderLen {
		return fmt.Errorf("expected to write %d bytes, wrote %d", expectedHeaderLen, n)
	}

	n, err = w.Write(b.Bytes)
	if err != nil {
		return err
	}
	if n != len(b.Bytes) {
		return fmt.Errorf("expected to write %d bytes, wrote %d", len(b.Bytes), n)
	}

	n, err = w.WriteString("\r\n")
	if err != nil {
		return err
	}

	if n != 2 {
		return fmt.Errorf("expected to write 2 bytes, wrote %d", n)
	}

	return w.Flush()
}

// BulkNullString represents a RESP Null Bulk String: $-1\r\n
type BulkNullString struct{}

func (BulkNullString) Write(w *bufio.Writer) error {
	n, err := w.WriteString("$-1\r\n")
	if err != nil {
		return err
	}

	if n != 5 {
		return fmt.Errorf("expected to write 5 bytes, wrote %d", n)
	}

	return w.Flush()
}

// Array represents a RESP Array. When Null is true, it encodes as *-1\r\n
type Array struct {
	Elems []Frame
	Null  bool
}

func (a Array) Write(w *bufio.Writer) error {
	if a.Null {
		n, err := w.WriteString("*-1\r\n")
		if err != nil {
			return err
		}
		if n != 4 {
			return fmt.Errorf("expected to write 4 bytes, wrote %d", n)
		}
		return err
	}
	n, err := fmt.Fprintf(w, "*%d\r\n", len(a.Elems))
	if err != nil {
		return err
	}
	// The header is "*<digits>\r\n" - we need to calculate actual digit length
	expectedHeaderLen := len(fmt.Sprintf("%d", len(a.Elems))) + 3 // * + digits + \r\n
	if n != expectedHeaderLen {
		return fmt.Errorf("expected to write %d bytes, wrote %d", expectedHeaderLen, n)
	}
	for _, el := range a.Elems {
		if err := el.Write(w); err != nil {
			return err
		}
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
		for range n {
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
