package world

import "fatewalker/game/item"

type World struct {
	Rooms map[string]*Room
	NPCs map[string][]*NPC
}

func NewWorld() *World {
	village := NewRoom("village_square", "Asterion Village Square", "Olive trees shade a quiet square in a small Greek village. Potters and fishmongers prepare for the day while the mountain road waits beyond the cottages.")
	lane := NewRoom("village_lane", "The Village Lane", "Whitewashed cottages and blue shutters line a narrow stone lane. A local guide waits beside a weathered milestone pointing toward the mountain.")
	gates := NewRoom("olympus_gates","The Gates of Olympus","Massive marble gates rise before you. Beyond them, Mount Olympus disappears into clouds illuminated by divine light. A weathered altar stands nearby, its inscription worn almost smooth: 'Every life begins with a choice.'")
	foothills := NewRoom("olympus_foothills","The Foothills of Olympus","A steep mountain path winds upward through ancient stone and mist. Far above, thunder rolls across the peak. Feathers drift through the air though there is no bird in sight.")
	oracle := NewRoom("oracle_path","The Oracle Path","Cypress trees crowd a winding trail toward Delphi. The air smells of smoke, laurel, and prophecy. Somewhere ahead, a voice seems to whisper a name you do not remember having.")
	den := NewRoom("manticore_den","The Manticore Den","Broken columns and old offerings surround a dark cavern. Something large has disturbed the dust. Ancient claw marks cross a wall beside a half-erased image of a three-headed hound.")
	crossroads := NewRoom("modern_crossroads","The Crossroads of Athens","Neon signs glow beside ancient stones. Cars hiss over wet pavement while, somewhere above the city, an impossible thunderclap answers your arrival.")

	gates.Exits["north"] = foothills.ID
	village.Exits["north"] = lane.ID
	lane.Exits["south"] = village.ID
	lane.Exits["north"] = foothills.ID
	foothills.Exits["south"] = gates.ID
	foothills.Exits["west"] = lane.ID
	foothills.Exits["north"]=oracle.ID
	oracle.Exits["south"]=foothills.ID
	oracle.Exits["east"]=den.ID
	den.Exits["west"]=oracle.ID

	w:=&World{Rooms:map[string]*Room{
		village.ID:village, lane.ID:lane, gates.ID:gates, foothills.ID:foothills, oracle.ID:oracle, den.ID:den, crossroads.ID:crossroads,
	}}
	w.NPCs=map[string][]*NPC{}
	guide := NewNPC("village_guide", "Damon, Village Guide", "A patient local leans on a walking stick beside the milestone.", 1, 100, 5, 0)
	guide.AddDialogue([]string{"hello", "greeting"}, `Damon smiles. "Welcome to Asterion. Take your time; even heroes should learn where their feet are going."`)
	guide.AddDialogue([]string{"road", "mountain", "north"}, `Damon points up the lane. "When you are ready, the northern road leads to the foothills. A harpy has been troubling travelers there. Finish your first lessons, then follow the road."`)
	guide.AddDialogue([]string{"village", "home"}, `"Asterion is small, but it remembers everyone who passes through. The mountain has a longer memory."`)
	w.NPCs["village_lane"] = append(w.NPCs["village_lane"], guide)
	oracleNPC:=NewNPC("pythia","Pythia","The Oracle of Delphi sits beside a brazier of fragrant smoke. Her eyes are closed, but she seems to have been waiting for you.",5,120,8,0)
	oracleNPC.AddDialogue([]string{"hello","greeting"}, `Pythia opens her eyes. "You carry a thread that has already crossed the river."`)
	oracleNPC.AddDialogue([]string{"thread","fate"}, `"The black thread is not a chain," the Oracle whispers. "It is a memory of a choice you have not yet made."`)
	oracleNPC.AddDialogue([]string{"styx","death","rebirth"}, `"When you cross the Styx, you will lose much. But the river cannot drink what your soul refuses to surrender."`)
	oracleNPC.AddDialogue([]string{"olympus","gods"}, `"The gods will offer you power. The Fates will offer you consequences. Do not confuse the two."`)
	oracleNPC.AddDialogue([]string{"help","quest"}, `"Follow the road to Delphi. Then seek the river. The first life is only the beginning."`)
	oracleNPC.AddDialogue([]string{"trials","laurel"}, `Pythia holds out a laurel leaf. "A seer must learn that vision is not certainty. Walk the sanctum, listen to the silence, and carry this lesson beyond death."`)
	w.NPCs["oracle_path"]=append(w.NPCs["oracle_path"],oracleNPC)
	armorer:=NewNPC("hephaestus_apprentice","Theron, Hephaestus' Apprentice","A soot-streaked young smith works a bronze blade over a glowing forge. His eyes flick briefly to the black thread around your weapon.",4,100,7,0)
	armorer.Faction = "olympians"
	armorer.AddDialogue([]string{"hello","greeting"}, `Theron nods. "If that sword remembers you, traveler, perhaps I should make it worth remembering."`)
	armorer.AddDialogue([]string{"thread","fate"}, `"I've seen strange metal before. Never metal that seemed to know its owner's name."`)
	armorer.Shop=[]item.Item{{Name:"ash spear",Tier:item.Uncommon,Damage:16,Kind:"spear",Price:75},{Name:"bronze aegis",Tier:item.Uncommon,Armor:5,Kind:"armor",Price:90}}
	w.NPCs["olympus_foothills"]=append(w.NPCs["olympus_foothills"],armorer)
	merchant:=NewNPC("athens_vendor","Myrto","A modern Athens street vendor watches the crowds with an amused smile. Ancient coins hang from her stall beside things that absolutely should not exist.",2,80,5,0)
	merchant.Faction = "underworld"
	merchant.AddDialogue([]string{"hello","greeting"}, `Myrto smiles. "You look like someone who has been away from Athens for a very, very long time."`)
	merchant.AddDialogue([]string{"ancient","past"}, `"Don't ask me how I get the old things. Ask yourself why you recognize them."`)
	merchant.Shop=[]item.Item{{Name:"city knife",Tier:item.Common,Damage:14,Kind:"sword",Price:60},{Name:"reinforced jacket",Tier:item.Uncommon,Armor:4,Kind:"armor",Price:80}}
	w.NPCs["modern_crossroads"]=append(w.NPCs["modern_crossroads"],merchant)
	expandWorld(w.Rooms)
	return w
}

func (w *World) GetRoom(id string) *Room { return w.Rooms[id] }
