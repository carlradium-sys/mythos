package world

import "fatewalker/game/item"

type World struct {
	Rooms map[string]*Room
	NPCs map[string][]*NPC
}

func NewWorld() *World {
	village := NewRoom("village_square", "Asterion Village Square", "Olive trees shade Asterion square. At the well, black water reflects a starless sky beneath the morning sun.")
	lane := NewRoom("village_lane", "The Village Lane", "Whitewashed cottages and blue shutters line a narrow stone lane. A local guide waits beside a weathered milestone pointing toward the mountain.")
	smithy := NewRoom("therons_forge", "Theron's Forge", "A low stone workshop clings to the mountain road. Coals glow beneath a bronze anvil, and unfinished weapons hang from the rafters. The forge is warm, quiet, and separate from the road outside.")
	gates := NewRoom("olympus_gates","The Gates of Olympus","Massive marble gates rise before you. Beyond them, Mount Olympus disappears into clouds illuminated by divine light. A weathered altar stands nearby, its inscription worn almost smooth: 'Every life begins with a choice.'")
	foothills := NewRoom("olympus_foothills","The Foothills of Olympus","A steep mountain path winds upward through ancient stone and mist. Far above, thunder rolls across the peak. Feathers drift through the air though there is no bird in sight.")
	oracle := NewRoom("oracle_path","The Oracle Path","Cypress trees crowd a winding trail toward Delphi. The air smells of smoke, laurel, and prophecy. Somewhere ahead, a voice seems to whisper a name you do not remember having.")
	den := NewRoom("manticore_den","The Manticore Den","Broken columns and old offerings surround a dark cavern. Something large has disturbed the dust. Ancient claw marks cross a wall beside a half-erased image of a three-headed hound.")
	crossroads := NewRoom("modern_crossroads","The Crossroads of Athens","Neon signs glow beside ancient stones. Cars hiss over wet pavement while, somewhere above the city, an impossible thunderclap answers your arrival.")

	// Isolated, one-way prologue rooms for newly created characters.
	tutShore := NewRoom("tutorial_shore", "The Shore Between Heartbeats", "You wake on pale sand beside a still pool of black water. The sky holds no sun, yet a warm dawn glows behind the hills. A narrow path leads north toward a lantern-lit shelter. Your hands are empty of memory, but a dark thread circles your wrist.")
	tutCamp := NewRoom("tutorial_camp", "The Wayfarer's Shelter", "A weathered shelter stands beneath olive branches. A guide has prepared a table with a map, a pack, and a practice blade. Nothing here is dangerous; this is a place to learn before the world begins testing you.")
	tutYard := NewRoom("tutorial_yard", "The Training Courtyard", "A low stone wall encloses a courtyard marked with chalk circles. A small training shade waits in the center, motionless and watchful. It will not strike first. Beyond the northern arch, a quiet grove offers a more realistic lesson.")
	tutGrove := NewRoom("tutorial_grove", "The Quiet Practice Grove", "Olive trees shade a ring of worn stones. A young satyr stands in the clearing with its staff lowered. It watches you but will not attack unless you attack it first. This is a controlled lesson, not a real ambush.")
	tutStore := NewRoom("tutorial_storehouse", "The Wayfarer's Storehouse", "A tidy storehouse holds a rack for weapons and a chest of practice supplies. Labels explain how carried items differ from equipped items. A marble gate to the north leads out of the lesson and into the living world.")
	tutGate := NewRoom("tutorial_gate", "The Gate of First Steps", "The tutorial ends at a pale marble gate. Beyond it, the road descends toward Asterion Village and the first true choices of your life. This threshold opens only once. Once you leave, the way back will be sealed.")
	tutShore.Exits["north"] = tutCamp.ID
	tutCamp.Exits["south"] = tutShore.ID
	tutCamp.Exits["east"] = tutYard.ID
	tutYard.Exits["west"] = tutCamp.ID
	tutYard.Exits["north"] = tutGrove.ID
	tutYard.Exits["east"] = tutStore.ID
	tutGrove.Exits["south"] = tutYard.ID
	tutStore.Exits["west"] = tutYard.ID
	tutStore.Exits["north"] = tutGate.ID
	tutGate.Exits["south"] = tutStore.ID
	tutGate.Exits["north"] = village.ID

	gates.Exits["north"] = foothills.ID
	village.Exits["north"] = lane.ID
	lane.Exits["south"] = village.ID
	lane.Exits["north"] = foothills.ID
	foothills.Exits["south"] = gates.ID
	foothills.Exits["west"] = lane.ID
	foothills.Exits["east"] = smithy.ID
	smithy.Exits["west"] = foothills.ID
	foothills.Exits["north"]=oracle.ID
	oracle.Exits["south"]=foothills.ID
	oracle.Exits["east"]=den.ID
	den.Exits["west"]=oracle.ID

	w:=&World{Rooms:map[string]*Room{
		village.ID:village, lane.ID:lane, smithy.ID:smithy, gates.ID:gates, foothills.ID:foothills, oracle.ID:oracle, den.ID:den, crossroads.ID:crossroads,
		tutShore.ID:tutShore, tutCamp.ID:tutCamp, tutYard.ID:tutYard, tutGrove.ID:tutGrove, tutStore.ID:tutStore, tutGate.ID:tutGate,
	}}
	w.NPCs=map[string][]*NPC{}
	tutGuide := NewNPC("tutorial_guide", "Ianthe, Wayfarer Guide", "A calm traveler in a faded blue cloak. She speaks as if she has guided many souls through their first uncertain steps.", 1, 100, 0, 0)
	tutGuide.AddDialogue([]string{"hello", "greeting", "name"}, `Ianthe smiles gently. "Welcome, wayfarer. You do not need to know everything at once. I will teach you one skill at a time. Read each room, try each command, and remember that help is always available."`)
	tutGuide.AddDialogue([]string{"movement", "move", "directions", "exits"}, `"Movement follows named exits. Use north, south, east, west, up, or down; short forms such as n and e work too. 'exits' lists routes and 'map' shows how nearby rooms connect. Looking at a path does not move you."`)
	tutGuide.AddDialogue([]string{"combat", "fight", "enemy", "monster"}, `"Inspect a creature before attacking. Some beings are hostile; others are wary but peaceful until harmed. 'attack <name>' starts a fight. Watch both health totals, use 'rest' only when safe, and remember that 'flee' is a legitimate survival choice."`)
	tutGuide.AddDialogue([]string{"inventory", "items", "equipment", "gear"}, `"Inventory is what you carry; equipment is what you have ready. Use 'inv' to inspect your pack and 'eq' to see your weapon and armor. Carrying a sword does not always mean it is the sword you have equipped."`)
	tutGuide.AddDialogue([]string{"story", "quests", "journal", "help"}, `"The journal gives broad story direction; quests track objectives. 'help <topic>' explains a system in depth, and 'tutorial' repeats the current lesson. You may explore beyond the obvious road later. This prologue is a safe place to practice."`)
	w.NPCs["tutorial_camp"] = append(w.NPCs["tutorial_camp"], tutGuide)

	guide := NewNPC("village_guide", "Damon, Village Guide", "A patient local leans on a walking stick beside the milestone.", 1, 100, 5, 0)
	guide.AddDialogue([]string{"hello", "greeting"}, `Damon studies the black thread at your wrist. "Three mornings ago, the river left you on our shore with no pulse and no name. At dawn you stood up again. The village has been kind, but the mountain has noticed. Learn the roads here before you follow that thread."`)
	guide.AddDialogue([]string{"road", "mountain", "north"}, `Damon points beyond the cottages. "The Harpy has driven travelers off the mountain road. Theron keeps a forge just east of the foothills if you need better steel, but first make sure you know how to read a room, fight, and check your pack."`)
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
	w.NPCs["therons_forge"]=append(w.NPCs["therons_forge"],armorer)
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
