package session

import (
	"io"
	"net"
	"testing"

	"fatewalker/game/character"
	"fatewalker/world"
)

func newChoiceTestSession(t *testing.T) *Session {
	t.Helper()
	server, client := net.Pipe()
	t.Cleanup(func() {
		server.Close()
		client.Close()
	})
	go func() {
		_, _ = io.Copy(io.Discard, client)
	}()
	c := character.New("Testwalker")
	c.RoomID = "oracle_path"
	return &Session{Character: c, Conn: server, World: world.NewWorld()}
}

func TestOracleTrustChoiceLeavesPersistentOath(t *testing.T) {
	s := newChoiceTestSession(t)
	s.choose([]string{"trust"})

	if !s.Character.HasStoryFlag("oracle_choice_made") || !s.Character.HasStoryFlag("oracle_trust") {
		t.Fatal("trust choice flags were not recorded")
	}
	if got := s.Character.Reputation["delphi"]; got != 2 {
		t.Fatalf("Delphi reputation = %d, want 2", got)
	}
	if len(s.Character.Oaths) != 1 || len(s.Character.Memories) != 1 {
		t.Fatal("trust choice should leave an oath and a memory")
	}

	s.choose([]string{"defy"})
	if s.Character.HasStoryFlag("oracle_defied") {
		t.Fatal("a second choice should not overwrite the first")
	}
	if got := s.Character.Reputation["delphi"]; got != 2 {
		t.Fatalf("second choice changed reputation to %d, want 2", got)
	}
}

func TestOracleDefianceLeavesPersistentScar(t *testing.T) {
	s := newChoiceTestSession(t)
	s.choose([]string{"defy"})

	if !s.Character.HasStoryFlag("oracle_choice_made") || !s.Character.HasStoryFlag("oracle_defied") {
		t.Fatal("defiance choice flags were not recorded")
	}
	if got := s.Character.Reputation["delphi"]; got != -1 {
		t.Fatalf("Delphi reputation = %d, want -1", got)
	}
	if len(s.Character.Scars) != 1 || len(s.Character.Echoes) != 1 {
		t.Fatal("defiance should leave a scar and an echo")
	}
}
