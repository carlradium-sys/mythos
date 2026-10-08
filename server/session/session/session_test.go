package session

import (
	"io"
	"net"
	"testing"

	"fatewalker/game/character"
	"fatewalker/game/combat"
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

func TestDefeatEnemyAwardsQuestProgressForCombatKills(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_foothills"
	s.Enemy = &combat.Enemy{Name: "Harpy", HP: 0, XP: 70}
	s.defeatEnemy(false)

	if got := s.Character.Quests["black_thread"]; got != 1 {
		t.Fatalf("black_thread progress = %d, want 1", got)
	}
	if s.Character.Experience != 70 {
		t.Fatalf("experience = %d, want 70", s.Character.Experience)
	}
	if s.Enemy != nil {
		t.Fatal("defeated enemy should be cleared")
	}
}

func TestManticoreDropWorksWithStarterInventoryAndDoesNotDuplicate(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Level = 5
	s.Enemy = &combat.Enemy{Name: "Manticore", HP: 0, XP: 0}
	s.defeatEnemy(true)

	count := 0
	for _, owned := range s.Character.Inventory {
		if owned.Name == "manticore fang" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("manticore fang count = %d, want 1", count)
	}

	s.Enemy = &combat.Enemy{Name: "Manticore", HP: 0, XP: 0}
	s.defeatEnemy(true)
	count = 0
	for _, owned := range s.Character.Inventory {
		if owned.Name == "manticore fang" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("repeated Manticore kill duplicated fang; count = %d", count)
	}
}
