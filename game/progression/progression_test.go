package progression

import "testing"

func TestLevelProgression(t *testing.T) {
	if got:=LevelFromXP(0); got!=1 { t.Fatalf("expected level 1, got %d",got) }
	if got:=LevelFromXP(XPForLevel(5)); got!=5 { t.Fatalf("expected level 5, got %d",got) }
	if got:=XPToNextLevel(1,0); got!=100 { t.Fatalf("expected 100 XP to level 2, got %d",got) }
}
