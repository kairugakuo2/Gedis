package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func main() {
	fmt.Println("Listening on port 6379!")

	// Create a new server:
	// net.Listen opens TCP socket on port 6379 + accepts connections
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Listener has failed to start: ", err)
		return
	}
	defer listener.Close()

	// Create or Open database Append Only File
	aof, err := NewAof("database.aof")
	if err != nil {
		fmt.Println("Error creating or opening database Aof", err)
		return
	}
	defer aof.Close()

	if err := aof.Read(func(value Value) {
		if value.typ != "array" || len(value.array) == 0 {
			fmt.Println("Skipping malformed entry in AOF")
			return
		}

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command in AOF: ", command)
			return
		}

		handler(args)
	}); err != nil {
		fmt.Println("Error replaying AOF: ", err)
		return
	}

	// Accept connections forever, one goroutine per client
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection failed to accept: ", err)
			continue
		}

		go handleConnection(connection, aof)
	}
}

func handleConnection(connection net.Conn, aof *Aof) {
	defer connection.Close()

	resp := NewResp(connection)
	writer := NewWriter(connection)

	for {
		value, err := resp.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client has disconnected")
			} else {
				fmt.Println("Error reading from client: ", err)
			}
			return
		}

		if value.typ != "array" || len(value.array) == 0 {
			writer.Write(Value{typ: "error", str: "ERR expected non-empty array"})
			continue
		}

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			writer.Write(Value{typ: "error", str: "ERR unknown command '" + command + "'"})
			continue
		}

		result := handler(args)

		// Only persist writes that actually succeeded
		if result.typ != "error" && (command == "SET" || command == "HSET") {
			if err := aof.Write(value); err != nil {
				fmt.Println("Error writing to AOF: ", err)
			}
		}

		writer.Write(result)
	}
}
