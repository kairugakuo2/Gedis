package main

import (
	"fmt"
	"io"
	"net"
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
	defer listener.Close() // run when main exits with defer (like 'using' in C# for cleanup)

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
		if readErr != nil {
			// io.EOF -> cleint hangs up cleanly
			if readErr == io.EOF {
				fmt.Println("Client has disconnected")
				break
			}
			fmt.Println("Error reading from client: ", readErr)
			break
		}

		fmt.Println(value)

		// give writer a Value and let it serialize to RESP
		writeErr := writer.Write(Value{typ: "string", str: "OK"})
		if writeErr != nil {
			fmt.Println("Error writing to client: ", writeErr)
			break
		}
	}

}
