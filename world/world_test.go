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


func TestTheronsForgeIsSeparateFromTheHarpyEncounter(t *testing.T) {
	w := NewWorld()
	foothills := w.GetRoom("olympus_foothills")
	forge := w.GetRoom("therons_forge")
	if foothills == nil || forge == nil {
		t.Fatal("foothills and Theron's forge must both exist")
	}
	if foothills.Exits["east"] != "therons_forge" || forge.Exits["west"] != "olympus_foothills" {
		t.Fatal("forge must have a two-way connection east of the foothills")
	}
	if len(w.NPCs["olympus_foothills"]) != 0 {
		t.Fatal("the Harpy encounter room should not also contain the apprentice")
	}
	if len(w.NPCs["therons_forge"]) != 1 || w.NPCs["therons_forge"][0].ID != "hephaestus_apprentice" {
		t.Fatal("Theron should be the forge's sole NPC")
	}
}

func TestMapNamesCurrentAndFutureLocations(t *testing.T) {
	w := NewWorld()
	got := w.MapText("delphi_sanctum")
	for _, want := range []string{"FATEWALKER WORLD MAP", "ANCIENT", "Asterion Village Square", "The Village Lane", "Theron's Forge", "The Gates of Olympus", "The Delphi Sanctum", "MODERN", "The Crossroads of Athens", "LATER LIVES", "The Lunar Oracle", ">@"} {
		if !strings.Contains(got, want) {
			t.Errorf("map output missing %q", want)
		}
	}
	if strings.Contains(got, "[ ]") {
		t.Fatal("map should use named locations rather than anonymous placeholders")
	}
}


func TestStoryNPCsExistAcrossEras(t *testing.T) {
	w := NewWorld()
	cases := []struct {
		room string
		id string
		keyword string
	}{
		{"styx_shore", "charon", "memory"},
		{"fields_asphodel", "asphodel_shade", "judgment"},
		{"modern_museum", "museum_curator", "artifact"},
		{"future_moon", "lunar_oracle", "rebirth"},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			var found *NPC
			for _, npc := range w.NPCs[tc.room] {
				if npc.ID == tc.id {
					found = npc
					break
				}
			}
			if found == nil {
				t.Fatalf("expected story NPC %q in room %q", tc.id, tc.room)
			}
			if _, text := found.MatchDialogue(tc.keyword); text == "" {
				t.Errorf("NPC %q has no dialogue for keyword %q", tc.id, tc.keyword)
			}
		})
	}
}


func TestLoreTopicsResolveThroughNPCDialogue(t *testing.T) {
	w := NewWorld()
	cases := []struct {
		room string
		npcID string
		topic string
	}{
		{"village_lane", "village_guide", "well"},
		{"village_lane", "village_guide", "harpy"},
		{"village_lane", "village_guide", "death"},
		{"oracle_path", "pythia", "thread"},
		{"therons_forge", "hephaestus_apprentice", "craft"},
		{"styx_shore", "charon", "fates"},
		{"fields_asphodel", "asphodel_shade", "token"},
		{"modern_museum", "museum_curator", "metro"},
		{"future_moon", "lunar_oracle", "shore"},
	}
	for _, tc := range cases {
		t.Run(tc.npcID+"/"+tc.topic, func(t *testing.T) {
			var found *NPC
			for _, npc := range w.NPCs[tc.room] {
				if npc.ID == tc.npcID {
					found = npc
					break
				}
			}
			if found == nil {
				t.Fatalf("NPC %q not found in room %q", tc.npcID, tc.room)
			}
			_, response := found.MatchDialogue(tc.topic)
			if response == "" {
				t.Fatalf("topic %q should resolve to story dialogue for %q", tc.topic, tc.npcID)
			}
			if len(strings.Fields(response)) < 25 {
				t.Errorf("lore response for %q/%q is too short to provide a meaningful story exchange", tc.npcID, tc.topic)
			}
		})
	}
}

func TestCoreLoreCluesAppearInWorldDescriptions(t *testing.T) {
	w := NewWorld()
	cases := []struct {
		room string
		clue string
	}{
		{"village_square", "bread and salt"},
		{"village_lane", "three tiny lines"},
		{"olympus_foothills", "coin pressed into the ash"},
		{"modern_crossroads", "slow lap of water"},
	}
	for _, tc := range cases {
		room := w.GetRoom(tc.room)
		if room == nil {
			t.Fatalf("expected lore room %q to exist", tc.room)
		}
		if !strings.Contains(strings.ToLower(room.Description), strings.ToLower(tc.clue)) {
			t.Errorf("room %q description should include recurring lore clue %q", tc.room, tc.clue)
		}
	}
}


func TestFoothillsHasInspectableCoinInAsh(t *testing.T) {
	w := NewWorld()
	r := w.GetRoom("olympus_foothills")
	if r == nil { t.Fatal("foothills room missing") }
	ash, coin := false, false
	for _, object := range r.Objects {
		if object.ID == "campfire_ash" && object.Container { ash = true }
		if object.ID == "scratched_bronze_coin" && object.ContainerID == "campfire_ash" && object.Takeable { coin = true }
	}
	if !ash || !coin { t.Fatal("foothills should contain a takeable bronze coin discoverable in campfire ash") }
}

func TestObjectKeywordsMatchNaturalNames(t *testing.T) {
	o := NewObject("coin", "scratched bronze coin", "A coin.", "coin", "token")
	if !o.Matches("bronze") || !o.Matches("token") { t.Fatal("object should match name fragments and keywords") }
}
