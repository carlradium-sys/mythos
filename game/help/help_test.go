package help

import (
	"strings"
	"testing"
)

func TestSpellNamesResolveToDetailedHelp(t *testing.T) {
	topic, ok := Get("Stormspark")
	if !ok {
		t.Fatal("Stormspark should have a help topic")
	}
	for _, expected := range []string{"DOMAIN: Storm", "UNLOCK: Level 5", "MANA COST: 8", "BASE DAMAGE: 30", "cast Stormspark", "ql add Stormspark"} {
		if !strings.Contains(topic.Text, expected) {
			t.Errorf("Stormspark help missing %q", expected)
		}
	}
}

func TestSpellHelpIgnoresSpacesAndCase(t *testing.T) {
	topic, ok := Get("judgment of the sky")
	if !ok || topic.Name != "Judgment of the Sky" {
		t.Fatalf("got topic %#v, ok=%v", topic, ok)
	}
}
