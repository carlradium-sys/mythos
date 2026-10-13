package help

import (
	"strings"
	"testing"
)

func TestEveryListedHelpTopicExists(t *testing.T) {
	for _, name := range Names() {
		if _, ok := Get(name); !ok {
			t.Errorf("Names lists %q but Get cannot find it", name)
		}
	}
}

func TestCoreTopicsAreDetailed(t *testing.T) {
	for _, name := range []string{"start", "movement", "combat", "guard", "progression", "economy", "world", "status", "invoke", "crafting", "consumables", "tutorial", "current"} {
		topic, ok := Get(name)
		if !ok {
			t.Errorf("missing core help topic %q", name)
			continue
		}
		if len(topic.Text) < 250 {
			t.Errorf("help topic %q is too brief: %d characters", name, len(topic.Text))
		}
	}
}

func TestSpellNamesResolveToDetailedHelp(t *testing.T) {
	topic, ok := Get("Stormspark")
	if !ok { t.Fatal("Stormspark should have a help topic") }
	for _, expected := range []string{"DOMAIN: Storm", "UNLOCK LEVEL: 5", "MANA COST: 8", "BASE DAMAGE: 30", "cast Stormspark", "ql set 1 Stormspark"} {
		if !strings.Contains(topic.Text, expected) { t.Errorf("Stormspark help missing %q", expected) }
	}
}

func TestSpellHelpMatchesCaseAndSpacing(t *testing.T) {
	topic, ok := Get("judgment of the sky")
	if !ok || topic.Name != "Judgment of the Sky" { t.Fatalf("got topic %#v, ok=%v", topic, ok) }
}


func TestCurrentQuestHelpExplainsTrackerWorkflow(t *testing.T) {
	topic, ok := Get("current")
	if !ok {
		t.Fatal("current help topic is missing")
	}
	for _, expected := range []string{"current quests", "current quests track <ID>", "current quest", "current quests untrack", "checklist", "next lead"} {
		if !strings.Contains(topic.Text, expected) {
			t.Errorf("current quest help missing %q", expected)
		}
	}
}


func TestListenHelpExplainsInteractionDrivenRiverProgress(t *testing.T) {
	topic, ok := Get("listen")
	if !ok {
		t.Fatal("listen help topic is missing")
	}
	for _, expected := range []string{"listen well", "listen water", "River of Memory", "not merely when you arrive"} {
		if !strings.Contains(topic.Text, expected) {
			t.Errorf("listen help missing %q", expected)
		}
	}
}


func TestTalkHelpExplainsOracleChoicePrerequisite(t *testing.T) {
	topic, ok := Get("talk")
	if !ok {
		t.Fatal("talk help topic is missing")
	}
	for _, expected := range []string{"talk Pythia what do you know about the thread", "Hear Pythia's warning", "not accept trust or defiance"} {
		if !strings.Contains(topic.Text, expected) {
			t.Errorf("talk help missing %q", expected)
		}
	}
}
