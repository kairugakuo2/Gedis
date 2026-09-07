package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Aof -> Append Only File -> redis records every command in Aof
// for persistence of data from memory -> disk
type Aof struct {
	file *os.File
	mu   sync.Mutex
	done chan struct{}
}

func NewAof(path string) (*Aof, error) {
	// create file if doesn't exist / open if it does
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		file: f,
		done: make(chan struct{}),
	}

	// start a goroutine to sync AOF to disk every 1 second while server running
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-aof.done:
				return
			case <-ticker.C:
				aof.mu.Lock()
				if err := aof.file.Sync(); err != nil {
					fmt.Println("Error syncing AOF: ", err)
				}
				aof.mu.Unlock()
			}
		}
	}()

	return aof, nil
}

// properly close file
func (aof *Aof) Close() error {
	close(aof.done)

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

	if _, err := aof.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	resp := NewResp(aof.file)

	for {
		value, err := resp.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		callback(value)
	}
}
