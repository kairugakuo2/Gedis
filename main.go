package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	fmt.Println("Lisening on port 6379!")

	// net.Listen opens TCP socket on port 6379 + accepts connections
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Listener has failed to start: ", err)
		return
	}

	// run when main exits with defer (like 'using' in C# for cleanup)
	defer listener.Close()

	// block until a client connects -> then can read from / write to client
	connection, err := listener.Accept()
	if err != nil {
		fmt.Println("Connection failed to accept: ", err)
		return
	}
	defer connection.Close()

	for {

		// allocate empty byte array
		buffer := make([]byte, 1024)

		// fill buffer with what client sent + return written byte count
		bytesRead, err := connection.Read(buffer)
		if err != nil {
			// io.EOF -> cleint hangs up cleanly
			if err == io.EOF {
				fmt.Println("Client has disconnected")
				break
			}
			fmt.Println("Error reading from client: ", err)
			return
		}

		fmt.Printf("Reciebed %d bytes: %q\n", bytesRead, buffer[:bytesRead])

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
