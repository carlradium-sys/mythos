package world

type World struct {
	Rooms map[string]*Room
}

func NewWorld() *World {
	gates := NewRoom("olympus_gates","The Gates of Olympus","Massive marble gates rise before you. Beyond them, Mount Olympus disappears into clouds illuminated by divine light. A weathered altar stands nearby, its inscription worn almost smooth: 'Every life begins with a choice.'")
	foothills := NewRoom("olympus_foothills","The Foothills of Olympus","A steep mountain path winds upward through ancient stone and mist. Far above, thunder rolls across the peak. Feathers drift through the air though there is no bird in sight.")
	oracle := NewRoom("oracle_path","The Oracle Path","Cypress trees crowd a winding trail toward Delphi. The air smells of smoke, laurel, and prophecy. Somewhere ahead, a voice seems to whisper a name you do not remember having.")
	den := NewRoom("manticore_den","The Manticore Den","Broken columns and old offerings surround a dark cavern. Something large has disturbed the dust. Ancient claw marks cross a wall beside a half-erased image of a three-headed hound.")
	crossroads := NewRoom("modern_crossroads","The Crossroads of Athens","Neon signs glow beside ancient stones. Cars hiss over wet pavement while, somewhere above the city, an impossible thunderclap answers your arrival.")

	gates.Exits["north"]=foothills.ID
	foothills.Exits["south"]=gates.ID
	foothills.Exits["north"]=oracle.ID
	oracle.Exits["south"]=foothills.ID
	oracle.Exits["east"]=den.ID
	den.Exits["west"]=oracle.ID

	return &World{Rooms:map[string]*Room{
		gates.ID:gates, foothills.ID:foothills, oracle.ID:oracle, den.ID:den, crossroads.ID:crossroads,
	}}
	expandWorld(w.Rooms)
	return w
}

func (w *World) GetRoom(id string) *Room { return w.Rooms[id] }
