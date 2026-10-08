package session

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"fatewalker/game/character"
	"fatewalker/world"
)

type Session struct {
	Character *character.Character
	Conn      net.Conn
	World     *world.World
}

func New(character *character.Character, conn net.Conn, w *world.World) *Session {
	return &Session{Character: character, Conn: conn, World: w}
}

func (s *Session) WriteLine(format string, args ...any) {
	fmt.Fprintf(s.Conn, format+"\r\n", args...)
}

func (s *Session) Run(scanner *bufio.Scanner) {
	s.WriteLine("")
	s.WriteLine("========================================")
	s.WriteLine("        FATEWALKER: BEYOND THE STYX")
	s.WriteLine("========================================")
	s.WriteLine("")
	s.WriteLine("Welcome, mortal.")
	s.WriteLine("Your first life begins in Ancient Greece.")
	s.WriteLine("")
	s.WriteLine("Type 'help' for commands.")
	s.WriteLine("")

	for {
		s.WriteLine("%s", s.prompt())
		if !scanner.Scan() {
			return
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if !s.handleCommand(input) {
			return
		}
	}
}

func (s *Session) prompt() string {
	return fmt.Sprintf("[%s L%d Life %d] > ", s.Character.Era, s.Character.Level, s.Character.Life)
}

func (s *Session) handleCommand(input string) bool {
	parts := strings.Fields(strings.ToLower(input))
	if len(parts) == 0 {
		return true
	}

	switch parts[0] {
	case "quit", "exit":
		s.WriteLine("Farewell, %s.", s.Character.Name)
		return false
	case "help", "?":
		s.help()
	case "look", "l":
		s.look()
	case "north", "south", "east", "west", "up", "down", "n", "s", "e", "w", "u", "d":
		s.move(parts[0])
	case "who":
		s.WriteLine("There are no other adventurers visible yet.")
	case "score", "stats":
		s.score()
	case "rebirth":
		s.rebirth()
	default:
		s.WriteLine("Unknown command. Type 'help' for help.")
	}
	return true
}

func (s *Session) help() {
	s.WriteLine("Commands:")
	s.WriteLine("  look (l)       Look around")
	s.WriteLine("  north/south    Move")
	s.WriteLine("  east/west      Move")
	s.WriteLine("  up/down        Move")
	s.WriteLine("  score           Character information")
	s.WriteLine("  rebirth         Begin a new incarnation when eligible")
	s.WriteLine("  who             See who is online")
	s.WriteLine("  help            Show this help")
	s.WriteLine("  quit            Leave Fatewalker")
}

func (s *Session) look() {
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil {
		s.WriteLine("You are nowhere. The world has lost track of you.")
		return
	}
	s.WriteLine("")
	s.WriteLine("%s", r.Name)
	s.WriteLine("%s", r.Description)
	if len(r.Exits) > 0 {
		exits := make([]string, 0, len(r.Exits))
		for direction := range r.Exits {
			exits = append(exits, direction)
		}
		s.WriteLine("Exits: %s", strings.Join(exits, ", "))
	}
}

func (s *Session) move(direction string) {
	aliases := map[string]string{"n":"north","s":"south","e":"east","w":"west","u":"up","d":"down"}
	if canonical, ok := aliases[direction]; ok {
		direction = canonical
	}
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil {
		s.WriteLine("You cannot find your way.")
		return
	}
	next, ok := r.Exits[direction]
	if !ok {
		s.WriteLine("You cannot go that way.")
		return
	}
	s.Character.RoomID = next
	s.look()
}

func (s *Session) score() {
	c := s.Character
	s.WriteLine("Name: %s", c.Name)
	s.WriteLine("Level: %d", c.Level)
	s.WriteLine("Life: %d", c.Life)
	s.WriteLine("Rebirths: %d", c.Rebirths)
	s.WriteLine("Era: %s", c.Era)
	s.WriteLine("HP: %d/%d", c.HP, c.MaxHP)
	s.WriteLine("Experience: %d", c.Experience)
}

func (s *Session) rebirth() {
	if !s.Character.CanRebirth() {
		s.WriteLine("The Fates have not yet opened the way. You must reach level 10.")
		return
	}
	oldLife := s.Character.Life
	s.Character.Rebirth()
	s.WriteLine("The world falls silent.")
	s.WriteLine("You cross the Styx.")
	s.WriteLine("Life %d ends. Life %d begins.", oldLife, s.Character.Life)
	s.WriteLine("Your incarnation is now aligned with the %s era.", s.Character.Era)
}
