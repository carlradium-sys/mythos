package combat

import (
	"strings"
	"testing"
)

func TestEnemyAttackNarrationIsDistinct(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{name: "Harpy", want: "Talons"},
		{name: "Satyr", want: "staff"},
		{name: "Manticore", want: "spines"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			approach, impact := enemyAttackNarration(tc.name)
			if !strings.Contains(approach, tc.name) {
				t.Fatalf("approach %q should name %s", approach, tc.name)
			}
			if !strings.Contains(impact, tc.want) {
				t.Fatalf("impact %q should contain %q", impact, tc.want)
			}
		})
	}
}

func TestUnknownEnemyHasFallbackCombatNarration(t *testing.T) {
	approach, impact := enemyAttackNarration("Gorgon")
	if !strings.Contains(approach, "Gorgon") || !strings.Contains(impact, "Gorgon") {
		t.Fatalf("fallback narration = %q / %q, want enemy name in both", approach, impact)
	}
}
