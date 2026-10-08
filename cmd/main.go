package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:4000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Fatewalker MUD server listening on 127.0.0.1:4000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("\r\n"))
	conn.Write([]byte("========================================\r\n"))
	conn.Write([]byte("          FATEWALKER: BEYOND THE STYX\r\n"))
	conn.Write([]byte("========================================\r\n"))
	conn.Write([]byte("\r\n"))
	conn.Write([]byte("Welcome, mortal.\r\n"))
	conn.Write([]byte("\r\n"))
	conn.Write([]byte("What is your name? "))

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			conn.Write([]byte("What is your name? "))
			continue
		}

		if strings.EqualFold(input, "quit") {
			conn.Write([]byte("Farewell, mortal.\r\n"))
			return
		}

		fmt.Fprintf(conn, "\r\nYou entered: %s\r\n", input)
		fmt.Fprintf(conn, "\r\nWhat is your name? ")
	}
}