package character

import (
	"encoding/json"
	"testing"
)

func TestStoryFlagsPersistThroughJSON(t *testing.T) {
	c := New("Ariadne")
	c.SetStoryFlag("oracle_choice_made")
	c.SetStoryFlag("oracle_trust")

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal character: %v", err)
	}
	var restored Character
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal character: %v", err)
	}
	if !restored.HasStoryFlag("oracle_choice_made") || !restored.HasStoryFlag("oracle_trust") {
		t.Fatal("story flags were not preserved by character persistence")
	}
	if restored.HasStoryFlag("oracle_defied") {
		t.Fatal("unselected choice should not be set")
	}
}

func TestSetStoryFlagInitializesLegacyNilMap(t *testing.T) {
	c := &Character{Name: "Legacy"}
	c.SetStoryFlag("old_soul")
	if !c.HasStoryFlag("old_soul") {
		t.Fatal("SetStoryFlag should initialize a nil map")
	}
}
