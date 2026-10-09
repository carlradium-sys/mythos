package world

import "fmt"

func (w *World) MapText(cur string) string {
	groups := []struct {
		name string
		ids  []string
	}{
		{"ANCIENT", []string{"village_square", "village_lane", "therons_forge", "olympus_gates", "olympus_foothills", "oracle_path", "delphi_sanctum", "manticore_den", "temple_dawn", "ancient_athens", "olympus_road", "olympus_hall", "olympus_garden", "styx_shore", "underworld_crossroads", "fields_asphodel", "hall_judgment", "cerberus_gate", "tartarus_edge"}},
		{"MODERN", []string{"modern_crossroads", "modern_plaka", "modern_acropolis", "modern_metro", "modern_museum", "modern_rooftop", "modern_styx", "modern_sanctum"}},
		{"LATER LIVES", []string{"future_city", "future_skyway", "future_moon", "far_era"}},
	}
	out := "\x1b[1;33mFATEWALKER WORLD MAP\x1b[0m\n"
	for _, group := range groups {
		out += fmt.Sprintf("\n\x1b[1;36m%s\x1b[0m\n", group.name)
		for _, id := range group.ids {
			room := w.GetRoom(id)
			if room == nil {
				continue
			}
			marker := "  "
			if id == cur {
				marker = "\x1b[1;36m>@\x1b[0m"
			}
			out += fmt.Sprintf("%s %s\n", marker, room.Name)
		}
	}
	out += "\n>@ marks your current location. Later eras become reachable through rebirth."
	return out
}
