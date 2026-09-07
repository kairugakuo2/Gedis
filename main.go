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

	aof.Read(func(value Value) {
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}

		handler(args)
	})
	// Listen for connections
	// block until a client connects -> then can read from / write to client
	connection, err := listener.Accept()
	if err != nil {
		fmt.Println("Connection failed to accept: ", err)
		return
	}
	defer connection.Close()

	resp := NewResp(connection)
	writer := NewWriter(connection)
	for {
		// parse client message into Value struct
		value, readErr := resp.Read()
		fmt.Println(value)
		if readErr != nil {
			// io.EOF -> cleint hangs up cleanly
			if readErr == io.EOF {
				fmt.Println("Client has disconnected")
				break
			}
			fmt.Println("Error reading from client: ", readErr)
			break
		}
		if value.typ != "array" {
			fmt.Println("Invalid request, expected array")
			continue
		}
		if len(value.array) == 0 {
			fmt.Println("Invalid request, expected array length > 0")
			continue
		}

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(Value{typ: "string", str: ""})
			continue
		}

		if command == "SET" || command == "HSET" {
			aof.Write(value)
		}

		result := handler(args)
		writer.Write(result)
	}

}
