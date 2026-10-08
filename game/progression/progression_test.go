package progression

import "testing"

func TestLevelProgression(t *testing.T) {
	if got:=LevelFromXP(0); got!=1 { t.Fatalf("expected level 1, got %d",got) }
	if got:=LevelFromXP(XPForLevel(5)); got!=5 { t.Fatalf("expected level 5, got %d",got) }
	if got:=XPToNextLevel(1,0); got!=60 { t.Fatalf("expected 60 XP to level 2, got %d",got) }
	if got:=XPForLevel(10); got!=4860 { t.Fatalf("expected level 10 rebirth threshold of 4860 XP, got %d",got) }
	if got:=LevelFromXP(4860); got!=10 { t.Fatalf("expected 4860 XP to reach level 10, got %d",got) }
}
