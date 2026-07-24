package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/NiravShah1729/semdb/protocol"
)

func main() {
	port := ":8080"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Echo server listening on %s...\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		fmt.Printf("Client connected from %s\n", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	r := protocol.NewReader(conn)
	w := protocol.NewWriter(conn)

	for {
		// 1. Read and parse incoming RESP message
		val, err := r.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected.")
				return
			}
			fmt.Println("Read error:", err)
			return
		}

		fmt.Printf("Received: %s\n", val)

		// 2. Prepare the response Value
		var response protocol.Value

		// Check if the payload is an Array with "PING" as the first argument
		if val.Type == protocol.TypeArray && len(val.Array) > 0 {
			cmd := strings.ToUpper(string(val.Array[0].Bulk))
			if cmd == "PING" {
				response = protocol.Value{
					Type: protocol.TypeSimpleString,
					Str:  "PONG",
				}
			} else {
				// Echo back the received Array
				response = val
			}
		} else if val.Type == protocol.TypeSimpleString && strings.ToUpper(val.Str) == "PING" {
			// Handle simple string +PING
			response = protocol.Value{
				Type: protocol.TypeSimpleString,
				Str:  "PONG",
			}
		} else {
			// Echo back any other parsed Value struct as-is
			response = val
		}

		// 3. Serialize and write the response using protocol.Writer
		if err := w.Write(response); err != nil {
			fmt.Println("Write error:", err)
			return
		}
	}
}