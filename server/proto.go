package server

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

type Request struct {
	Command       string
	Arguments     [][]byte
	RemoteAddress string
	Connection    io.ReadCloser
}

func (r *Request) HasArgument(index int) bool {
	return index >= 0 && index < len(r.Arguments)
}

func (r *Request) ExpectArgument(index int) *ErrorReply {
	if !r.HasArgument(index) {
		return ErrNotEnoughArgs
	}
	return nil
}

func (r *Request) GetInt(index int) (int64, *ErrorReply) {
	if errReply := r.ExpectArgument(index); errReply != nil {
		return -1, errReply
	}
	if n, err := strconv.ParseInt(string(r.Arguments[index]), 10, 64); err != nil {
		return -1, ErrExpectInteger
	} else {
		return n, nil
	}
}

func NewRequest(conn io.ReadCloser) (*Request, error) {
	reader := bufio.NewReader(conn)

	// *<number of arguments>CRLF
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	var argCount int
	if line[0] == '*' {
		if _, err := fmt.Sscanf(line, "*%d\r\n", &argCount); err != nil {
			return nil, Malformed("*<#Arguments>", line)
		}

		// $<number of bytes of argument 1>CRLF
		// <argument data>CRLF
		command, err := readArgument(reader)
		if err != nil {
			return nil, err
		}

		arguments := make([][]byte, argCount-1)
		for i := 0; i < argCount-1; i++ {
			if arguments[i], err = readArgument(reader); err != nil {
				return nil, err
			}
		}

		return &Request{
			Command:    strings.ToUpper(string(command)),
			Arguments:  arguments,
			Connection: conn,
		}, nil
	}

	return nil, fmt.Errorf("new request error")
}

func readArgument(reader *bufio.Reader) ([]byte, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, Malformed("$<ArgumentLength>", line)
	}

	var argLength int
	if _, err := fmt.Sscanf(line, "$%d\r\n", &argLength); err != nil {
		return nil, Malformed("$<ArgumentLength>", line)
	}

	data, err := ioutil.ReadAll(io.LimitReader(reader, int64(argLength)))
	if err != nil {
		return nil, err
	}
	if len(data) != argLength {
		return nil, MalformedLength(argLength, len(data))
	}
	if b, err := reader.ReadByte(); err != nil || b != '\r' {
		return nil, MalformedMissingCRLF()
	}
	if b, err := reader.ReadByte(); err != nil || b != '\n' {
		return nil, MalformedMissingCRLF()
	}

	return data, nil
}

func Malformed(expected string, got string) error {
	return fmt.Errorf("Mailformed request: %s does not match %s", got, expected)
}

func MalformedLength(expected int, got int) error {
	return fmt.Errorf("Mailformed request: argument length %d does not match %d", got, expected)
}

func MalformedMissingCRLF() error {
	return fmt.Errorf("Mailformed request: line should end with CRLF")
}

type Reply io.WriterTo

var (
	ErrMethodNotSupported   = &ErrorReply{"Method is not supported"}
	ErrNotEnoughArgs        = &ErrorReply{"Not enough arguments for the command"}
	ErrTooMuchArgs          = &ErrorReply{"Too many arguments for the command"}
	ErrWrongArgsNumber      = &ErrorReply{"Wrong number of arguments"}
	ErrExpectInteger        = &ErrorReply{"Expected integer"}
	ErrExpectPositivInteger = &ErrorReply{"Expected positive integer"}
	ErrExpectMorePair       = &ErrorReply{"Expected at least one key val pair"}
	ErrExpectEvenPair       = &ErrorReply{"Got uneven number of key val pairs"}

	ErrNoKey = &ErrorReply{"no key for set"}
)

type ErrorReply struct {
	message string
}

func (er *ErrorReply) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte("-ERROR " + er.message + "\r\n"))
	return int64(n), err
}

func (er *ErrorReply) Error() string {
	return er.message
}

type StatusReply struct {
	code string
}

func (r *StatusReply) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte("+" + r.code + "\r\n"))
	return int64(n), err
}

type IntReply struct {
	number int64
}

func (r *IntReply) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(":" + strconv.FormatInt(r.number, 10) + "\r\n"))
	return int64(n), err
}

type BulkReply struct {
	value []byte
}

func (r *BulkReply) WriteTo(w io.Writer) (int64, error) {
	return writeBytes(r.value, w)
}

type MultiBulkReply struct {
	values [][]byte
}

func (r *MultiBulkReply) WriteTo(w io.Writer) (int64, error) {
	if r.values == nil {
		return 0, fmt.Errorf("Multi bulk reply found a nil values")
	}
	if wrote, err := w.Write([]byte("*" + strconv.Itoa(len(r.values)) + "\r\n")); err != nil {
		return int64(wrote), err
	} else {
		total := int64(wrote)
		for _, value := range r.values {
			wroteData, err := writeBytes(value, w)
			total += wroteData
			if err != nil {
				return total, err
			}
		}
		return total, nil
	}
}

func writeNullBytes(w io.Writer) (int64, error) {
	n, err := w.Write([]byte("$-1\r\n"))
	return int64(n), err
}

func writeBytes(value interface{}, w io.Writer) (int64, error) {
	if value == nil {
		return writeNullBytes(w)
	}
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			return writeNullBytes(w)
		}
		buf := []byte("$" + strconv.Itoa(len(v)) + "\r\n")
		buf = append(buf, v...)
		buf = append(buf, []byte("\r\n")...)
		n, err := w.Write(buf)
		if err != nil {
			return 0, err
		}
		return int64(n), nil
	case int:
		wrote, err := w.Write([]byte(":" + strconv.Itoa(v) + "\r\n"))
		return int64(wrote), err
	}
	return 0, fmt.Errorf("Invalid type sent to WriteBytes")
}
