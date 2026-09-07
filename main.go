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
	for {
		// parse client message into Value struct
		value, err := resp.Read()
		if err != nil {
			// io.EOF -> cleint hangs up cleanly
			if err == io.EOF {
				fmt.Println("Client has disconnected")
				break
			}
			fmt.Println("Error reading from client: ", err)
			break
		}

		fmt.Println(value)

		// reply wih a RESP simple string:
		// '+' , the text, a carriave return, and newline
		// currently ignoring wha client asked for and always say OK

		_, err = connection.Write([]byte("+OK\r\n"))
		if err != nil {
			fmt.Println("Error writing to client: ", err)
			break
		}
	}

}
