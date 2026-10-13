package world

import "fatewalker/game/item"

type World struct {
	Rooms map[string]*Room
	NPCs map[string][]*NPC
}

func NewWorld() *World {
	village := NewRoom("village_square", "Asterion Village Square", "The square has been awake for hours. Olive leaves whisper above the well, where the water is so dark that it reflects no face—only a starless sky, though the morning sun shines overhead. A woman draws her bucket without looking down. Beside the shrine, yesterday’s offerings have been swept into a neat pile, except for one black feather caught beneath a stone. The baker’s door stands open, filling the square with the smell of warm barley bread. For a village this small, everyone seems determined not to look at the well.")
	lane := NewRoom("village_lane", "The Village Lane", "The lane bends between whitewashed cottages, their blue shutters faded by salt wind and summer sun. Someone has hung garlic above a doorway; someone else has left a child’s wooden horse on the step. At the far end, a weather-beaten milestone points toward the mountain road. Damon, the village guide, leans against it with the patience of a man who has already answered the same question too many times today. When he sees you, his expression changes—not to surprise, exactly, but recognition he would rather not explain.")
	smithy := NewRoom("therons_forge", "Theron's Forge", "The forge has been built into the hillside, its stone walls blackened by years of smoke. Coals pulse beneath a bronze anvil, and unfinished spearheads hang from the rafters like dull moons. The air tastes of charcoal, hot metal, and rain on stone. Theron studies a blade as though it has said something offensive. At the sound of your steps, he turns it so its edge catches the light.")
	gates := NewRoom("olympus_gates","The Gates of Olympus","Massive marble gates rise before you. Beyond them, Mount Olympus disappears into clouds illuminated by divine light. A weathered altar stands nearby, its inscription worn almost smooth: 'Every life begins with a choice.'")
	foothills := NewRoom("olympus_foothills","The Foothills of Olympus","The mountain road climbs through broken terraces and old retaining walls, their stones fitted so tightly that grass has nowhere to grow. Mist gathers in the hollows, carrying the smell of wet rock and something faintly singed. Thunder moves along the summit, too slow and deliberate to be weather. Black feathers drift across the path. One has lodged in a crack in the stone; its quill trembles although the air around it is still. Far below, the village bells begin to ring, then stop after a single note.")
	oracle := NewRoom("oracle_path","The Oracle Path","Cypress trees crowd a winding trail toward Delphi. The air smells of smoke, laurel, and prophecy. Somewhere ahead, a voice seems to whisper a name you do not remember having.")
	den := NewRoom("manticore_den","The Manticore Den","Broken columns and old offerings surround a dark cavern. Something large has disturbed the dust. Ancient claw marks cross a wall beside a half-erased image of a three-headed hound.")
	crossroads := NewRoom("modern_crossroads","The Crossroads of Athens","Neon signs glow beside ancient stones. Cars hiss over wet pavement while, somewhere above the city, an impossible thunderclap answers your arrival.")

	// Isolated, one-way prologue rooms for newly created characters.
	tutShore := NewRoom("tutorial_shore", "The Shore Between Heartbeats", "Cold sand presses against your cheek. Beyond it, black water lies perfectly still, reflecting a sky without sun or stars. A warm light gathers behind the distant hills, like dawn remembered by someone who has never seen it. You cannot recall falling, or walking, or the name by which anyone might have called you. A dark thread circles your wrist; when you touch it, something far away seems to draw a breath. North, a lantern burns beneath the roof of a small shelter. A voice carries from it, gentle but not surprised: “You are awake, then.”")
	tutCamp := NewRoom("tutorial_camp", "The Wayfarer's Shelter", "Olive branches shelter a low stone building whose roof has been repaired with three kinds of tile. Inside, a lamp burns beside a rough table. Someone has laid out a folded map, a travel pack, a waterskin, and a practice blade—not arranged like a lesson, but like the belongings of a traveler expected to leave soon. Ianthe sits near the doorway, mending a tear in her blue cloak. She does not ask who you are. Instead, she studies the black thread at your wrist and quietly moves the cup beside her to make room.")
	tutYard := NewRoom("tutorial_yard", "The Training Courtyard", "A low stone wall encloses a courtyard marked with chalk circles. A small training shade waits in the center, motionless and watchful. It will not strike first. Beyond the northern arch, a quiet grove offers a more realistic lesson.")
	tutGrove := NewRoom("tutorial_grove", "The Quiet Practice Grove", "Olive trees shade a ring of worn stones. A young satyr stands in the clearing with its staff lowered. It watches you but will not attack unless you attack it first. This is a controlled lesson, not a real ambush.")
	tutStore := NewRoom("tutorial_storehouse", "The Wayfarer's Storehouse", "The storehouse is cool and smells of cedar, oil, and old dust. A rack holds practice weapons worn smooth by many hands; a chest beneath it contains bandages, simple supplies, and sandals repaired with blue thread. There are no instructions nailed to the walls, only a charcoal sketch of a traveler’s pack beside a sketch of a shield. Through the northern doorway, pale marble catches the daylight. Beyond that gate lies a village that has never heard of this shelter—and a road that will not wait forever.")
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
	village.Items = []item.Item{{Name:"black feather", Tier:item.Common, Kind:"clue"}, {Name:"cracked clay token", Tier:item.Common, Kind:"keepsake"}}
	tutCamp.Items = []item.Item{{Name:"folded travel map", Tier:item.Common, Kind:"tool"}, {Name:"clay waterskin", Tier:item.Common, Kind:"consumable", Healing:5}}
	tutStore.Items = []item.Item{{Name:"linen bandage", Tier:item.Common, Kind:"consumable", Healing:12}, {Name:"bronze oil lamp", Tier:item.Common, Kind:"tool"}}
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
	armorer.Shop=[]item.Item{{Name:"ash spear",Tier:item.Uncommon,Damage:16,Kind:"spear",Price:75},{Name:"bronze aegis",Tier:item.Uncommon,Armor:5,Kind:"armor",Price:90},{Name:"ambrosia draught",Tier:item.Uncommon,Kind:"consumable",Price:35,Healing:45},{Name:"blue lotus tonic",Tier:item.Common,Kind:"consumable",Price:25,ManaRestore:20},{Name:"laurel poultice",Tier:item.Common,Kind:"consumable",Price:18,Healing:20}}
	w.NPCs["therons_forge"]=append(w.NPCs["therons_forge"],armorer)
	merchant:=NewNPC("athens_vendor","Myrto","A modern Athens street vendor watches the crowds with an amused smile. Ancient coins hang from her stall beside things that absolutely should not exist.",2,80,5,0)
	merchant.Faction = "underworld"
	merchant.AddDialogue([]string{"hello","greeting"}, `Myrto smiles. "You look like someone who has been away from Athens for a very, very long time."`)
	merchant.AddDialogue([]string{"ancient","past"}, `"Don't ask me how I get the old things. Ask yourself why you recognize them."`)
	merchant.Shop=[]item.Item{{Name:"city knife",Tier:item.Common,Damage:14,Kind:"sword",Price:60},{Name:"reinforced jacket",Tier:item.Uncommon,Armor:4,Kind:"armor",Price:80},{Name:"ambrosia draught",Tier:item.Uncommon,Kind:"consumable",Price:40,Healing:45},{Name:"blue lotus tonic",Tier:item.Common,Kind:"consumable",Price:28,ManaRestore:20},{Name:"laurel poultice",Tier:item.Common,Kind:"consumable",Price:20,Healing:20}}
	w.NPCs["modern_crossroads"]=append(w.NPCs["modern_crossroads"],merchant)
	expandWorld(w.Rooms)
	return w
}

func (w *World) GetRoom(id string) *Room { return w.Rooms[id] }
