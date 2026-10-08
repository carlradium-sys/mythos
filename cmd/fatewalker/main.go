package main

import (
	"bufio"
	"fmt"
	"net"

	"fatewalker/game/character"
	"fatewalker/server/session/session"
	"fatewalker/world"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:4000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Fatewalker: Beyond the Styx listening on 127.0.0.1:4000")

	w := world.NewWorld()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go handleConnection(conn, w)
	}
}

func handleConnection(conn net.Conn, w *world.World) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	c := character.New("Mortal")
	s := session.New(c, conn, w)
	s.Run(scanner)
}
