package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"fatewalker/game/character"
	"fatewalker/world"
)

type Player struct {
	Character *character.Character
	Conn      net.Conn
}

var (
	players   = make(map[net.Conn]*Player)
	playersMu sync.RWMutex

	gameWorld = world.NewWorld()
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:4000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("========================================")
	fmt.Println("       FATEWALKER: BEYOND THE STYX")
	fmt.Println("========================================")
	fmt.Println("Server listening on 127.0.0.1:4000")
	fmt.Println()

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

	send(conn, "")
	send(conn, "========================================")
	send(conn, "          FATEWALKER: BEYOND THE STYX")
	send(conn, "========================================")
	send(conn, "")
	send(conn, "Welcome, mortal.")
	send(conn, "")
	send(conn, "What is your name? ")

	scanner := bufio.NewScanner(conn)

	if !scanner.Scan() {
		return
	}

	name := strings.TrimSpace(scanner.Text())

	if name == "" {
		send(conn, "Invalid name.")
		return
	}

	if playerNameExists(name) {
		send(conn, "")
		send(conn, "That character is already in the world.")
		send(conn, "Please choose another name.")
		send(conn, "")
		return
	}

	player := &Player{
		Character: character.New(name),
		Conn:      conn,
	}

	playersMu.Lock()
	players[conn] = player
	playersMu.Unlock()

	send(conn, "")
	send(conn, "Welcome, "+name+".")
	send(conn, "")
	send(conn, "You stand at the gates of a world ruled by gods.")
	send(conn, "")
	send(conn, "Type 'help' for commands.")
	send(conn, "")

	broadcast(name+" has entered the world.", conn)

	send(conn, "> ")

	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			send(conn, "> ")
			continue
		}

		if strings.EqualFold(input, "quit") {
			send(conn, "")
			send(conn, "Farewell, "+player.Character.Name+".")
			send(conn, "")

			removePlayer(conn)
			return
		}

		handleCommand(conn, input)

		send(conn, "> ")
	}

	removePlayer(conn)

}

func handleCommand(conn net.Conn, input string) {
	switch strings.ToLower(input) {
	case "look":
		look(conn)

	case "north", "n":
		movePlayer(conn, "north")

	case "south", "s":
		movePlayer(conn, "south")

	case "who":
		showPlayers(conn)

	case "help":
		showHelp(conn)

	default:
		send(conn, "Unknown command: "+input)
	}

}

func movePlayer(conn net.Conn, direction string) {
	playersMu.RLock()
	player := players[conn]
	playersMu.RUnlock()

	if player == nil {
		return
	}

	currentRoom := gameWorld.GetRoom(player.Character.RoomID)

	if currentRoom == nil {
		send(conn, "You are nowhere.")
		return
	}

	destinationID, exists := currentRoom.Exits[direction]

	if !exists {
		send(conn, "You cannot go that way.")
		return
	}

	destination := gameWorld.GetRoom(destinationID)

	if destination == nil {
		send(conn, "That path leads nowhere.")
		return
	}

	player.Character.RoomID = destination.ID

	send(conn, "")
	send(conn, "You travel "+direction+".")
	send(conn, "")

	look(conn)

}

func look(conn net.Conn) {
	playersMu.RLock()
	player := players[conn]
	playersMu.RUnlock()

	if player == nil {
		return
	}

	room := gameWorld.GetRoom(player.Character.RoomID)

	if room == nil {
		send(conn, "You are nowhere.")
		return
	}

	send(conn, "")
	send(conn, room.Name)
	send(conn, "========================================")
	send(conn, room.Description)
	send(conn, "")

	if len(room.Exits) > 0 {
		var exits []string

		for direction := range room.Exits {
			exits = append(exits, direction)
		}

		send(conn, "Exits: "+strings.Join(exits, ", "))
	} else {
		send(conn, "Exits: none")
	}

	send(conn, "")

}

func showHelp(conn net.Conn) {
	send(conn, "")
	send(conn, "Available commands:")
	send(conn, "----------------------------------------")
	send(conn, " look - Look around")
	send(conn, " north - Travel north")
	send(conn, " south - Travel south")
	send(conn, " n - Travel north")
	send(conn, " s - Travel south")
	send(conn, " who - See who is online")
	send(conn, " help - Show this help")
	send(conn, " quit - Disconnect")
	send(conn, "")
}

func showPlayers(conn net.Conn) {
	playersMu.RLock()
	defer playersMu.RUnlock()

	send(conn, "")
	send(conn, "Players currently online:")
	send(conn, "----------------------------------------")

	if len(players) == 0 {
		send(conn, "Nobody.")
		send(conn, "")
		return
	}

	for _, player := range players {
		send(conn, "  "+player.Character.Name)
	}

	send(conn, "")

}

func playerNameExists(name string) bool {
	playersMu.RLock()
	defer playersMu.RUnlock()

	for _, player := range players {
		if strings.EqualFold(player.Character.Name, name) {
			return true
		}
	}

	return false

}

func broadcast(message string, except net.Conn) {
	playersMu.RLock()
	defer playersMu.RUnlock()

	for _, player := range players {
		if player.Conn != except {
			send(player.Conn, message)
		}
	}

}

func removePlayer(conn net.Conn) {
	playersMu.Lock()

	player, exists := players[conn]

	if exists {
		delete(players, conn)
	}

	playersMu.Unlock()

	if exists {
		broadcast(player.Character.Name+" has left the world.", nil)
	}

}

func send(conn net.Conn, message string) {
	fmt.Fprintf(conn, "%s\r\n", message)
}
