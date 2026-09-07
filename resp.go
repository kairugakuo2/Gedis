package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// constants to represent data types we're working with
const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

// parse / deserialize RESP commands from client
type Value struct {
	typ   string  // value
	str   string  // simple sring
	num   int     // integer
	bulk  string  // bulk strings
	array []Value // arrays
}

// reader -> to deserialize RESP
type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
	// to pass buffer from the Connection
	return &Resp{reader: bufio.NewReader(rd)}
}

// readLine -> read line from buffer
func (r *Resp) readLine() (line []byte, n int, err error) {
	// read one byte at a time until '\r' -> end of line
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1

		line = append(line, b)

		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}

	// return w/out last 2 bytes -> '\r\n' + num of bytes
	return line[:len(line)-2], n, nil
}

// readInteger -> read integer from buffer
func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	num, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(num), n, nil
}

// Read() -> reccursivley read bufer

func (r *Resp) Read() (Value, error) {
	// read first byte -> RESP type to parse
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	// parse according to type
	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknwon type: %v", string(_type))
		return Value{}, nil
	}
}

func (r *Resp) readArray() (Value, error) {
	v := Value{}
	v.typ = "array"

	// read length of array
	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	// foreach line, parse + read value
	v.array = make([]Value, length)
	for i := 0; i < length; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		// add parsed value to array
		v.array[i] = val
	}

	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.typ = "bulk"

	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, length)

	r.reader.Read(bulk)

	v.bulk = string(bulk)

	// Read the trailing CRLF -> '\r\n'
	r.readLine()

	return v, nil
}
