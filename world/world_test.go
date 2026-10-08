package world

import "testing"

func TestNewWorldHasNoDanglingExits(t *testing.T) {
	w := NewWorld()
	if len(w.Rooms) == 0 {
		t.Fatal("NewWorld returned an empty room map")
	}
	for id, room := range w.Rooms {
		if room == nil {
			t.Errorf("room %q is nil", id)
			continue
		}
		for direction, target := range room.Exits {
			if _, ok := w.Rooms[target]; !ok {
				t.Errorf("room %q has %q exit to missing room %q", id, direction, target)
			}
		}
	}
}

func TestNewWorldContainsTempleOfTheFirstDawn(t *testing.T) {
	w := NewWorld()
	if _, ok := w.Rooms["temple_dawn"]; !ok {
		t.Fatal("Temple of the First Dawn must exist because it connects Delphi to the Styx")
	}
}
