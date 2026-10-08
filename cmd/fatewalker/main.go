package main

import (
	"bufio"
	"fmt"
	"net"

	"fatewalker/game/account"
	"fatewalker/server/session/session"
	"fatewalker/world"
)

func main() {
	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Fatewalker: Beyond the Styx listening on :4000")

	w := world.NewWorld()
	accounts := account.NewStore("data/accounts")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go handleConnection(conn, w, accounts)
	}
}

func handleConnection(conn net.Conn, w *world.World, accounts *account.Store) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	s := session.New(accounts, conn, w)
	s.Run(scanner)
}
