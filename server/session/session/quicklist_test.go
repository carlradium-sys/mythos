package session

import (
	"testing"

	"fatewalker/game/combat"
	"fatewalker/game/power"
)

func TestQuicklistAddsAndCastsUnlockedPowerAtNamedTarget(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Level = 5
	s.Character.Domain = power.Storm
	s.Enemy = &combat.Enemy{Name: "Harpy", HP: 100, MaxHP: 100, Level: 1, XP: 10}

	s.quicklist([]string{"add", "stormspark"})
	if len(s.Character.Quicklist) != 1 || s.Character.Quicklist[0] != "Stormspark" {
		t.Fatalf("quicklist = %#v, want [Stormspark]", s.Character.Quicklist)
	}

	manaBefore := s.Character.Mana
	hpBefore := s.Enemy.HP
	s.quicklist([]string{"1", "satyr"})
	if s.Character.Mana != manaBefore || s.Enemy.HP != hpBefore {
		t.Fatal("invalid target should not cast or consume mana")
	}

	s.quicklist([]string{"1", "harpy"})
	if s.Character.Mana != manaBefore-8 {
		t.Fatalf("mana = %d, want %d after Stormspark", s.Character.Mana, manaBefore-8)
	}
	if s.Enemy != nil && s.Enemy.HP >= hpBefore {
		t.Fatalf("enemy HP did not decrease after quicklist cast: %d", s.Enemy.HP)
	}
}

func TestQuicklistRejectsLockedPower(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Level = 5
	s.Character.Domain = power.Storm

	s.quicklist([]string{"add", "Thunderfall"})
	if len(s.Character.Quicklist) != 0 {
		t.Fatalf("locked power was added: %#v", s.Character.Quicklist)
	}
}
