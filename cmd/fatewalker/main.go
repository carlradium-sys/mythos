package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"fatewalker/game/account"
	"fatewalker/server/session/session"
	"fatewalker/world"
)

func main() {
	addr := os.Getenv("FATEWALKER_ADDR")
	if addr == "" {
		addr = "127.0.0.1:4000"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Printf("Fatewalker: Beyond the Styx listening on %s\n", addr)

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
