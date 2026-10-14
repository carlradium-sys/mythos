package character

import (
	"encoding/json"
	"testing"

	"fatewalker/game/item"
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

func TestStoryFlagsSurviveRebirth(t *testing.T) {
	c := New("Threadbound")
	c.SetStoryFlag("oracle_choice_made")
	c.SetStoryFlag("oracle_trust")
	c.RebirthTo("modern", "modern_crossroads")

	if !c.HasStoryFlag("oracle_choice_made") || !c.HasStoryFlag("oracle_trust") {
		t.Fatal("story choices should remain part of the soul after rebirth")
	}
	if c.Life != 2 || c.Era != "modern" {
		t.Fatalf("rebirth state = life %d, era %q; want life 2, modern", c.Life, c.Era)
	}
}


func TestRebirthGrantsDestinationSpecificLifeGift(t *testing.T) {
	c := New("GiftBearer")
	if c.LifeGift != "Thread Sense" {
		t.Fatalf("first life gift = %q, want Thread Sense", c.LifeGift)
	}
	c.LifeGiftUsed = true
	c.RebirthTo("modern", "modern_crossroads")
	if c.LifeGift != "Echo Sight" {
		t.Fatalf("modern life gift = %q, want Echo Sight", c.LifeGift)
	}
	if c.LifeGiftUsed {
		t.Fatal("life gift should reset on rebirth")
	}
	c.RebirthTo("future", "future_moon")
	if c.LifeGift != "Moon's Shelter" {
		t.Fatalf("lunar life gift = %q, want Moon's Shelter", c.LifeGift)
	}
	c.RebirthTo("far", "far_era")
	if c.LifeGift != "Fateweaver's Knot" {
		t.Fatalf("far life gift = %q, want Fateweaver's Knot", c.LifeGift)
	}
}


func TestEnsureLifeGiftMigratesOlderCharacter(t *testing.T) {
	c := New("LegacySoul")
	c.Era = "modern"
	c.RoomID = "modern_crossroads"
	c.LifeGift = ""
	c.LifeGiftUsed = true
	c.EnsureLifeGift()
	if c.LifeGift != "Echo Sight" {
		t.Fatalf("migrated life gift = %q, want Echo Sight", c.LifeGift)
	}
	if c.LifeGiftUsed {
		t.Fatal("newly migrated life-gift should be available")
	}
}


func TestCombatPreferencesAndQuicklistPersistThroughJSON(t *testing.T) {
	c := New("Quickcaster")
	c.AutoCombat = false
	c.CombatModeSet = true
	c.Quicklist = [4]string{"Stormspark", "Thunderfall", "", "Judgment of the Sky"}

	data, err := json.Marshal(c)
	if err != nil { t.Fatalf("marshal character: %v", err) }
	var restored Character
	if err := json.Unmarshal(data, &restored); err != nil { t.Fatalf("unmarshal character: %v", err) }
	if restored.AutoCombat || !restored.CombatModeSet {
		t.Fatal("manual combat preference was not persisted")
	}
	if restored.Quicklist != c.Quicklist {
		t.Fatalf("quicklist after persistence = %#v, want %#v", restored.Quicklist, c.Quicklist)
	}
}


func TestContainerStoragePersistsThroughJSON(t *testing.T) {
	c := New("Keeper")
	c.ContainerStorage["tutorial_storehouse/tutorial_practice_chest"] = []item.Item{
		{Name: "practice bandage", Kind: "consumable", Healing: 20},
	}
	data, err := json.Marshal(c)
	if err != nil { t.Fatalf("marshal character: %v", err) }
	var restored Character
	if err := json.Unmarshal(data, &restored); err != nil { t.Fatalf("unmarshal character: %v", err) }
	stored := restored.ContainerStorage["tutorial_storehouse/tutorial_practice_chest"]
	if len(stored) != 1 || stored[0].Healing != 20 || stored[0].Name != "practice bandage" {
		t.Fatalf("stored item did not survive persistence: %#v", stored)
	}
}
