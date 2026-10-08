package world

import ("testing"; "strings")

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


func TestMapNamesCurrentAndFutureLocations(t *testing.T) {
	w := NewWorld()
	got := w.MapText("delphi_sanctum")
	for _, want := range []string{"FATEWALKER WORLD MAP", "ANCIENT", "The Gates of Olympus", "The Delphi Sanctum", "MODERN", "The Crossroads of Athens", "LATER LIVES", "The Lunar Oracle", ">@"} {
		if !strings.Contains(got, want) {
			t.Errorf("map output missing %q", want)
		}
	}
	if strings.Contains(got, "[ ]") {
		t.Fatal("map should use named locations rather than anonymous placeholders")
	}
}
