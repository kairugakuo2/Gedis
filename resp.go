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

const (
	maxBulkSize = 512 * 1024 * 1024 // 512MB, same as Redis
	maxArrayLen = 1024 * 1024
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
		return Value{}, fmt.Errorf("unknown RESP type: %q", string(_type))
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

	if length < 0 {
		return Value{typ: "null"}, nil
	}
	if length > maxArrayLen {
		return v, fmt.Errorf("array too long: %d", length)
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

	if length < 0 {
		return Value{typ: "null"}, nil
	}
	if length > maxBulkSize {
		return v, fmt.Errorf("bulk string too large: %d", length)
	}

	bulk := make([]byte, length)
	if _, err := io.ReadFull(r.reader, bulk); err != nil {
		return v, err
	}

	v.bulk = string(bulk)

	// Read the trailing CRLF -> '\r\n'
	if _, _, err := r.readLine(); err != nil {
		return v, err
	}

	return v, nil
}

// Marshal will convert the Value to bytes (representing RESP response)
func (v Value) Marshal() []byte {
	// check and call method by type
	switch v.typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshalNull()
	case "error":
		return v.marshalError()
	default:
		return []byte{}
	}
}
func (v Value) marshalArray() []byte {
	var bytes []byte
	length := len(v.array)
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(length)...)
	bytes = append(bytes, '\r', '\n')

	// recursivley call marshal on Value to convert (regarless of type)
	for i := 0; i < length; i++ {
		bytes = append(bytes, v.array[i].Marshal()...)
	}

	return bytes
}
func (v Value) marshalBulk() []byte {
	// create byte array, add string, add CRLF ('\r\n')
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}
func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

// Null and Error incase data not found / error
func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}
func (v Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

// Writer -> write bytes from marshal to writer
type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

// take Value and write bytes from marshal to Writer
func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}
