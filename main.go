package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
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

	scanner := bufio.NewScanner(conn)

	remoteAddr := conn.RemoteAddr().String()
	for scanner.Scan() {
		text := scanner.Text()

		//echo 
		_,err := fmt.Fprintf(conn,"Echo: %s\n",text)
		if err != nil{
			log.Println("Counld not write back to the client")
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Print("Error reading from client")
	}

	fmt.Printf("Client disconnected: %s\n",remoteAddr)
} 