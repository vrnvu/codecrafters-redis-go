package protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}

	if len(line) < 2 || !strings.HasSuffix(line, "\r\n") {
		return "", fmt.Errorf("malformed line")
	}

	return line[:len(line)-2], nil // strip \r\n
}

func readBulk(r *bufio.Reader) (string, error) {
	line, err := readLine(r)
	if err != nil || !strings.HasPrefix(line, "$") {
		return "", fmt.Errorf("invalid bulk header")
	}

	n, err := strconv.Atoi(line[1:])
	if err != nil || n < 0 {
		return "", fmt.Errorf("invalid bulk length")
	}

	buf := make([]byte, n+2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}

	if !bytes.HasSuffix(buf, []byte("\r\n")) {
		return "", fmt.Errorf("bulk data missing CRLF")
	}

	return string(buf[:n]), nil
}

func ReadArray(r *bufio.Reader) ([]string, error) {
	line, err := readLine(r)
	if err != nil || !strings.HasPrefix(line, "*") {
		return nil, fmt.Errorf("invalid array header")
	}

	n, err := strconv.Atoi(line[1:])
	if err != nil || n < 0 {
		return nil, fmt.Errorf("invalid array length")
	}

	parts := make([]string, 0, n)
	for range n {
		part, err := readBulk(r)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	return parts, nil
}

func WriteSimpleString(conn net.Conn, msg string) {
	conn.Write([]byte("+" + msg + "\r\n"))
}

func WriteError(conn net.Conn, msg string) {
	conn.Write([]byte("-ERR " + msg + "\r\n"))
}
