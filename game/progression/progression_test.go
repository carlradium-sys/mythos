package progression

import "testing"

func TestLevelProgression(t *testing.T) {
	if got:=LevelFromXP(0); got!=1 { t.Fatalf("expected level 1, got %d",got) }
	if got:=LevelFromXP(XPForLevel(5)); got!=5 { t.Fatalf("expected level 5, got %d",got) }
	if got:=XPToNextLevel(1,0); got!=45 { t.Fatalf("expected 45 XP to level 2, got %d",got) }
	if got:=XPForLevel(10); got!=3645 { t.Fatalf("expected level 10 rebirth threshold of 3645 XP, got %d",got) }
	if got:=LevelFromXP(3645); got!=10 { t.Fatalf("expected 3645 XP to reach level 10, got %d",got) }
}
