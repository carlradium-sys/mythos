package world

type World struct {
	Rooms map[string]*Room
}

func NewWorld() *World {
	gates := NewRoom(
		"olympus_gates",
		"The Gates of Olympus",
		"Massive marble gates rise before you. Beyond them, Mount Olympus disappears into clouds illuminated by divine light.",
	)

	foothills := NewRoom(
		"olympus_foothills",
		"The Foothills of Olympus",
		"A steep mountain path winds upward through ancient stone and mist. Far above, thunder rolls across the peak.",
	)

	gates.Exits["north"] = foothills.ID
	foothills.Exits["south"] = gates.ID

	return &World{
		Rooms: map[string]*Room{
			gates.ID:     gates,
			foothills.ID: foothills,
		},
	}

}

func (w *World) GetRoom(id string) *Room {
	return w.Rooms[id]
}
