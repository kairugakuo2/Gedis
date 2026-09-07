package main

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"
)

// Aof -> Append Only File -> redis records every command in Aof
// for persistence of data from memory -> disk
type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
}

func NewAof(path string) (*Aof, error) {
	// create file if doesn't exist / open if it does
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	// create reader to read from file
	aof := &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}

	// start a goroutine to sync AOF to disk every 1 second while server running
	go func() {
		for {
			aof.mu.Lock()

			aof.file.Sync()

			aof.mu.Unlock()

			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

// properly close file
func (aof *Aof) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.file.Close()
}

// write command to Aof whenever request is recieved from client
func (aof *Aof) Write(value Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Write(value.Marshal())
	if err != nil {
		return err
	}

	return nil
}

// read command for Aof
func (aof *Aof) Read(callback func(value Value)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	resp := NewResp(aof.file)

	for {
		value, err := resp.Read()
		if err == nil {
			callback(value)
		}
		if err == io.EOF {
			break
		}
		return err
	}

	return nil
}
