package main

import (
	"fmt"
	"log"
	"net"
	"io"

	"github.com/NiravShah1729/semdb/protocol"
)

func main() {
	listener, err := net.Listen("tcp",":8080")

	if err != nil {
		log.Fatalf("Failed to start the server: %v",err)
	}

	defer listener.Close()
	fmt.Println("Echo server listning...")

	for {
		conn,err := listener.Accept()
		fmt.Printf("This is what conn looks like %v",conn)
		if err != nil {
			log.Printf("Failed to accept connection %v",err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn){
	defer conn.Close()

	r := protocol.NewReader(conn)

	for {
		// 1. Parse incoming RESP command using your Reader
		val, err := r.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected.")
				return
			}
			fmt.Println("Read error:", err)
			return
		}

		// Print what your parser extracted in the server console
		fmt.Printf("Received: %s\n", val)

		// 2. Send a raw RESP response back over the TCP socket
		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Write error:", err)
			return
		}
	}
} 