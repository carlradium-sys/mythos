package session

import (
	"net"

	"fatewalker/game/character"
)

type Session struct {
	Character *character.Character
	Conn      net.Conn
}

func New(character *character.Character, conn net.Conn) *Session {
	return &Session{
		Character: character,
		Conn:       conn,
	}
}
