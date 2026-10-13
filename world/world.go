package world

import "fatewalker/game/item"

type World struct {
	Rooms map[string]*Room
	NPCs map[string][]*NPC
}

func NewWorld() *World {
	village := NewRoom("village_square", "Asterion Village Square", "The square of Asterion was built around an olive tree split by lightning, though no storm has touched this valley in living memory. Beneath its roots, the village well holds water dark as polished obsidian. The elders say the well was here before the first house and that, on certain mornings, it reflects not the sky above but a night that has not yet happened. Fresh offerings of bread and salt line its rim. No one admits to leaving them.")
	lane := NewRoom("village_lane", "The Village Lane", "Whitewashed cottages lean close over a lane worn smooth by generations of bare feet, funeral processions, and returning soldiers. Blue shutters hang open, but conversation falters whenever the mountain comes into view. Beside a milestone, Damon waits with the patience of a man who has watched too many travelers leave and too few return. Someone has scratched three tiny lines into the stone: one straight, one broken, one unfinished.")
	smithy := NewRoom("therons_forge", "Theron's Forge", "A low stone workshop clings to the mountain road. Coals glow beneath a bronze anvil, and unfinished weapons hang from the rafters. The forge is warm, quiet, and separate from the road outside.")
	gates := NewRoom("olympus_gates","The Gates of Olympus","Massive marble gates rise before you. Beyond them, Mount Olympus disappears into clouds illuminated by divine light. A weathered altar stands nearby, its inscription worn almost smooth: 'Every life begins with a choice.'")
	foothills := NewRoom("olympus_foothills","The Foothills of Olympus","The road climbs through limestone terraces where old votive niches have been hacked from the rock and then deliberately sealed. Thunder travels across Olympus although the sky is clear. Black feathers tumble against the wind. At the edge of the path, a traveler’s sandals have been arranged neatly beside a cold campfire; there is no sign of the traveler, only a bronze coin pressed into the ash with its face scratched away. The mountain does not feel abandoned. It feels attentive.")
	oracle := NewRoom("oracle_path","The Oracle Path","Cypress trees crowd a winding trail toward Delphi, their roots gripping stones older than the road. Laurel smoke drifts downhill against the breeze. Travelers leave small gifts at the roadside shrines: a lock of hair, a cracked cup, a name written on linen and folded until the letters can no longer be read. Somewhere ahead, a voice speaks your name in the tone of someone correcting a mistake. You cannot tell whether it comes from the sanctuary or from a memory you have not lived yet.")
	den := NewRoom("manticore_den","The Manticore Den","Broken columns and old offerings surround a dark cavern. Something large has disturbed the dust. Ancient claw marks cross a wall beside a half-erased image of a three-headed hound.")
	crossroads := NewRoom("modern_crossroads","The Crossroads of Athens","Athens has paved over its old roads but never quite buried them. Neon spills across rain-dark stone; scooters thread between buses; a kiosk radio argues about football beneath a statue whose inscription has been replaced three times. For one second, every phone on the street displays the same unfamiliar date. Then the screens correct themselves. Beneath the traffic noise comes the slow lap of water against wood, impossibly close and impossibly deep. A vendor watches you as if she heard it too.")

	// Objects are world fixtures, not prose-only decoration. Their state is
	// changed by get/take commands and remains changed for the shared world.
	ashBed := NewObject("campfire_ash", "campfire ash", "A cold fire has collapsed into a shallow bed of grey ash. Someone has stirred it recently; a small round edge of bronze shows beneath the surface.", "ash", "ashes", "fire", "campfire")
	ashBed.Container = true
	ashBed.Takeable = false
	coin := NewObject("scratched_bronze_coin", "scratched bronze coin", "A bronze coin lies half-buried in the ash, still faintly warm despite the cold fire. One face has been scratched away. On the other, a tiny mark resembles a thread tied in a knot.", "coin", "bronze coin", "token")
	coin.ContainerID = ashBed.ID
	coin.Takeable = true
	foothills.Objects = append(foothills.Objects, ashBed, coin)

	// A sealed coffer rewards players who inspect, open, search, and read rather
	// than treating every room description as a checklist.
	coffer := NewObject("tomb_coffer", "sealed votive coffer", "A squat cedar coffer is bound with blackened bronze. The lock has no keyhole; three shallow marks are pressed into its lid: a straight line, a broken line, and an unfinished one.", "coffer", "chest", "box", "votive chest")
	coffer.Container = true
	coffer.Open = false
	coffer.Takeable = false
	prayer := NewObject("charred_prayer_strip", "charred prayer strip", "The surviving words are written in an older hand: 'I did not ask to be remembered. I asked that someone remember why I stayed.' Beneath the sentence, the same three-line mark has been drawn with a trembling hand.", "prayer", "strip", "prayer strip", "writing", "parchment")
	prayer.ContainerID = coffer.ID
	prayer.Takeable = true
	den.Objects = append(den.Objects, coffer, prayer)

	offeringBowl := NewObject("offering_bowl", "empty offering bowl", "A shallow clay bowl rests beside the old well. A thumbprint remains in the dust inside it, though no one admits to having touched it.", "bowl", "clay bowl", "offering")
	offeringBowl.Takeable = false
	village.Objects = append(village.Objects, offeringBowl)

	shrineLamp := NewObject("bronze_oil_lamp", "bronze oil lamp", "The lamp is tarnished where many hands have held it. Its wick is cold, but the faint scent of myrrh clings to the bronze.", "lamp", "oil lamp", "bronze lamp")
	shrineLamp.Takeable = true
	foothills.Objects = append(foothills.Objects, shrineLamp)

	// Isolated, one-way prologue rooms for newly created characters.
	tutShore := NewRoom("tutorial_shore", "The Shore Between Heartbeats", "You wake on pale sand beside a still pool of black water. The sky holds no sun, yet a warm dawn glows behind the hills. A narrow path leads north toward a lantern-lit shelter. Your hands are empty of memory, but a dark thread circles your wrist.")
	tutCamp := NewRoom("tutorial_camp", "The Wayfarer's Shelter", "A weathered shelter stands beneath olive branches. A guide has prepared a table with a map, a pack, and a practice blade. Nothing here is dangerous; this is a place to learn before the world begins testing you.")
	tutYard := NewRoom("tutorial_yard", "The Training Courtyard", "A low stone wall encloses a courtyard marked with chalk circles. A small training shade waits in the center, motionless and watchful. It will not strike first. Beyond the northern arch, a quiet grove offers a more realistic lesson.")
	tutGrove := NewRoom("tutorial_grove", "The Quiet Practice Grove", "Olive trees shade a ring of worn stones. A young satyr stands in the clearing with its staff lowered. It watches you but will not attack unless you attack it first. This is a controlled lesson, not a real ambush.")
	tutStore := NewRoom("tutorial_storehouse", "The Wayfarer's Storehouse", "A tidy storehouse holds a rack for weapons and a chest of practice supplies. Labels explain how carried items differ from equipped items. A marble gate to the north leads out of the lesson and into the living world.")

	// The prologue's table and practice supplies are tangible, inspectable
	// objects so players can learn that room descriptions refer to real things.
	tutorialMap := NewObject("tutorial_map", "weathered practice map", "A charcoal map marks the shelter, the courtyard, the grove, and the one-way gate. The ink is smudged where someone traced the road north with a thumb.", "map", "charcoal map")
	tutorialMap.Takeable = false
	tutCamp.Objects = append(tutCamp.Objects, tutorialMap)
	practiceToken := NewObject("practice_token", "clay practice token", "A small clay disk stamped with a simple spiral. It has no value, but its weight makes it easier to understand how objects become part of your inventory.", "token", "clay disk", "practice coin")
	practiceToken.Takeable = true
	tutStore.Objects = append(tutStore.Objects, practiceToken)
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
	tutGuide.AddDialogue([]string{"hello", "greeting", "name"}, `Ianthe sets a cup of warm water on the table before asking your name. "You were found beside a shore that is not on any map. You were cold, but not dead—or perhaps you had been dead and something changed its mind. I have guided other lost souls, though none arrived with a thread around their wrist. You can stay here until you feel steady. Look around, ask what you need, and do not mistake my advice for orders. The road will still be there when you are ready."`)
	tutGuide.AddDialogue([]string{"movement", "move", "directions", "exits"}, `"Movement follows named exits. Use north, south, east, west, up, or down; short forms such as n and e work too. 'exits' lists routes and 'map' shows how nearby rooms connect. Looking at a path does not move you."`)
	tutGuide.AddDialogue([]string{"combat", "fight", "enemy", "monster"}, `"Inspect a creature before attacking. Some beings are hostile; others are wary but peaceful until harmed. 'attack <name>' starts a fight. Watch both health totals, use 'rest' only when safe, and remember that 'flee' is a legitimate survival choice."`)
	tutGuide.AddDialogue([]string{"inventory", "items", "equipment", "gear"}, `"Inventory is what you carry; equipment is what you have ready. Use 'inv' to inspect your pack and 'eq' to see your weapon and armor. Carrying a sword does not always mean it is the sword you have equipped."`)
	tutGuide.AddDialogue([]string{"objects", "items", "inspect", "examine", "coin", "map"}, `"A place is more than its walls and exits. The map on this table is a real thing you can inspect, just as the objects you find later will be real things you can pick up. Try 'look weathered practice map', then use 'iscan' when you want a list of things you can handle. 'scan' also lists people and creatures."`)
	tutGuide.AddDialogue([]string{"story", "quests", "journal", "help"}, `"The journal gives broad story direction; quests track objectives. 'help <topic>' explains a system in depth, and 'tutorial' repeats the current lesson. You may explore beyond the obvious road later. This prologue is a safe place to practice."`)
	w.NPCs["tutorial_camp"] = append(w.NPCs["tutorial_camp"], tutGuide)

	guide := NewNPC("village_guide", "Damon, Village Guide", "Damon's walking stick is olive wood, polished by years of use. His sandals are mended with different-colored thread, and he watches the mountain with the practiced caution of someone who has buried friends there.", 1, 100, 5, 0)
	guide.AddDialogue([]string{"hello", "greeting"}, `Damon studies the black thread at your wrist before he looks at your face. "Three mornings ago, a shepherd found you beside the old well. No breath. No heartbeat. No wound that could explain either. At dawn, you sat up and asked what year it was. When we told you, you wept—not because you remembered your death, but because you said we had got the year wrong." He rubs a thumb over the worn head of his stick. "Asterion has a rule for strange things: we offer bread first, questions second. But the mountain has been sending its creatures down the road, and they look at you as though you are late for an appointment. Learn where you are before you follow the thread."`)
	guide.AddDialogue([]string{"road", "mountain", "north"}, `Damon points beyond the cottages, but his eyes remain on the ridge. "The Harpy is not hunting goats. It tears open packs, scatters letters, and leaves the coins untouched. Last week it carried a wedding veil to the summit and pinned it to a shrine that has been sealed since my grandfather's grandfather was a boy. Theron keeps a forge east of the foothills; he can mend what the road breaks. But don't mistake a weapon for an answer. Look at the signs, learn how to defend yourself, and if the creature falls, search the place where it chose to make its nest. Monsters have habits. Sometimes habits are the only testimony the dead leave us."`)
	guide.AddDialogue([]string{"village", "home"}, `"Asterion is small enough that every grief becomes public and every kindness gets remembered. We bury our dead beneath olive trees because the roots keep the soil from washing away. There is an older custom, too: we leave an empty bowl at the table for the person who has not come home. My mother says the bowl is not an invitation. It is a promise that someone still knows there is a place missing."`)
	guide.AddDialogue([]string{"well", "water", "black"}, `Damon's expression closes. "The well was here before Asterion had a name. My grandmother said the first villagers found it full of seawater, though the sea is three days away. On moonless nights, the water reflects a shore none of us has visited. Once, a child looked down and saw her own funeral. She lived to be ninety. She never looked into it again. If you hear someone calling from the well, do not answer with your name."`)
	guide.AddDialogue([]string{"harpy", "feathers", "nest"}, `"The Harpy takes what people meant to carry home," Damon says. "Letters, keepsakes, wedding things. It leaves coin and food. My brother thinks it is collecting a nest; I think it is building a shrine from things people have not finished saying. If you find a message among its feathers, read it carefully. A monster can steal a letter without understanding it. The person who wrote it may have understood too much."`)
	guide.AddDialogue([]string{"death", "dead", "return"}, `Damon does not answer at once. "We have a custom for strangers who return from death: we do not ask what waited beyond it until they have eaten. Hunger makes every memory sound like a prophecy. Eat first. Then decide which parts of what you remember deserve to be believed."`)
	w.NPCs["village_lane"] = append(w.NPCs["village_lane"], guide)
	oracleNPC:=NewNPC("pythia","Pythia","The Oracle of Delphi sits beside a brazier of fragrant smoke. Her eyes are closed, but she seems to have been waiting for you.",5,120,8,0)
	oracleNPC.AddDialogue([]string{"hello","greeting"}, `Pythia opens her eyes. The smoke from her brazier bends toward you, though no wind moves in the chamber. "You carry a thread that has already crossed the river. I have seen it in the dreams of the dying and in the hands of children who have not yet been born. Neither vision showed me your face. Both showed me the same empty place at the table of the gods."`)
	oracleNPC.AddDialogue([]string{"thread","fate"}, `Pythia draws a black strand through the ash on the altar. "The Fates spin lives, but mortals are forever mistaking the cloth for a cage. This thread is not a chain. It is the memory of a choice you have not yet made—and the wound left by a choice someone made in your name. Do not ask whether your path is written. Ask who benefits from telling you that it cannot be changed."`)
	oracleNPC.AddDialogue([]string{"styx","death","rebirth"}, `"The Styx does not simply carry the dead," Pythia says. "It separates what happened from what can be borne. Some souls surrender their names because a name hurts; others surrender the face of the one they loved, believing pain will follow it into the dark. The river takes what is offered and what is forgotten. But a memory guarded with honesty can become a shore. You will cross it more than once. The danger is not that you will forget. It is that you will mistake what remains for the whole truth."`)
	oracleNPC.AddDialogue([]string{"olympus","gods"}, `"The gods are not one family with one purpose. They are powers forced to share a table: pride beside grief, hospitality beside vengeance, order beside appetite. They will offer you gifts and call the gift a bond. The Fates offer no gifts at all—only the consequences that remain after the feast is over. Take power if you must. Never let the giver name the price without reading the oath."`)
	oracleNPC.AddDialogue([]string{"help","quest"}, `"Follow the road to Delphi. Then seek the river. The first life is only the beginning."`)
	oracleNPC.AddDialogue([]string{"trials","laurel"}, `Pythia holds out a laurel leaf. "A seer must learn that vision is not certainty. Walk the sanctum, listen to the silence, and carry this lesson beyond death."`)
	w.NPCs["oracle_path"]=append(w.NPCs["oracle_path"],oracleNPC)
	armorer:=NewNPC("hephaestus_apprentice","Theron, Hephaestus' Apprentice","A soot-streaked young smith works a bronze blade over a glowing forge. His eyes flick briefly to the black thread around your weapon.",4,100,7,0)
	armorer.Faction = "olympians"
	armorer.AddDialogue([]string{"hello","greeting"}, `Theron lifts the blade from the coals and watches its reflection crawl along the bronze. "A sword should remember the hand that holds it, not the hand that killed the last owner. Yours has a strange habit of warming when the dead are near. I could call that a blessing, but blessings are what people name a danger before it asks for payment. Show me what the mountain leaves behind. I can turn a feather, horn, or spine into something that might keep you alive long enough to learn why your weapon knows you."`)
	armorer.AddDialogue([]string{"thread","fate"}, `"I've seen strange metal before. Never metal that seemed to know its owner's name. When you bring me a monster's remains, I don't ask whether the beast was evil. I ask what it did to survive, what it protected, and what part of it the world taught to become a weapon. Steel keeps no moral record. The hand that shapes it should."`)
	armorer.AddDialogue([]string{"coffer", "chest", "manticore", "den"}, `Theron wipes soot from his hands. "There is an old votive coffer in the den north of here. I once tried to pry it open with a chisel; the bronze simply bent back into shape. The lock is not mechanical. Three marks are pressed into the lid — a straight line, a broken one, and one that never finishes. If it opens for you, do not sell what you find. Some things are worth more as witnesses than as materials."`)
	armorer.AddDialogue([]string{"prayer", "witness", "name"}, `"The dead leave all sorts of things behind," Theron says. "Weapons, certainly. But also letters, apologies, recipes, a child's name scratched under a table. A name can be erased. The reason someone stayed is harder to destroy. If you find a prayer with no name, read it before you decide what it is worth."`)
	armorer.AddDialogue([]string{"forge", "craft", "materials"}, `Theron lays a feather beside a horn and a strip of blackened hide. "A Harpy feather holds the memory of wind; a Satyr horn keeps a stubborn note of music; a Manticore spine remembers every place it pierced. That is why monster-work is dangerous. You are not merely using a dead creature's strength. You are asking its last instinct to live on in your armor. A careless smith makes a curse. A careful one makes a warning you can wear."`)
	armorer.Shop=[]item.Item{{Name:"ash spear",Tier:item.Uncommon,Damage:16,Kind:"spear",Price:75},{Name:"bronze aegis",Tier:item.Uncommon,Armor:5,Kind:"armor",Price:90},{Name:"ambrosia draught",Tier:item.Uncommon,Kind:"consumable",Price:35,Healing:45},{Name:"blue lotus tonic",Tier:item.Common,Kind:"consumable",Price:25,ManaRestore:20},{Name:"laurel poultice",Tier:item.Common,Kind:"consumable",Price:18,Healing:20}}
	w.NPCs["therons_forge"]=append(w.NPCs["therons_forge"],armorer)
	merchant:=NewNPC("athens_vendor","Myrto","A modern Athens street vendor watches the crowds with an amused smile. Ancient coins hang from her stall beside things that absolutely should not exist.",2,80,5,0)
	merchant.Faction = "underworld"
	merchant.AddDialogue([]string{"hello","greeting"}, `Myrto smiles. "You look like someone who has been away from Athens for a very, very long time."`)
	merchant.AddDialogue([]string{"ancient","past"}, `"Don't ask me how I get the old things. Ask yourself why you recognize them."`)
	merchant.Shop=[]item.Item{{Name:"city knife",Tier:item.Common,Damage:14,Kind:"sword",Price:60},{Name:"reinforced jacket",Tier:item.Uncommon,Armor:4,Kind:"armor",Price:80},{Name:"ambrosia draught",Tier:item.Uncommon,Kind:"consumable",Price:40,Healing:45},{Name:"blue lotus tonic",Tier:item.Common,Kind:"consumable",Price:28,ManaRestore:20},{Name:"laurel poultice",Tier:item.Common,Kind:"consumable",Price:20,Healing:20}}
	w.NPCs["modern_crossroads"]=append(w.NPCs["modern_crossroads"],merchant)
	expandWorld(w.Rooms)

	// These clues turn major story milestones into interactions, not room-entry
	// checkboxes. They are ordinary room objects and can be examined by name.
	if sanctum := w.Rooms["delphi_sanctum"]; sanctum != nil {
		inscription := NewObject("delphi_broken_inscription", "broken stone inscription", "Most of the lettering has been scraped away. What remains describes a soul that returned from the river carrying a memory no living witness could confirm. The final line has been deliberately chiseled out.", "inscription", "stone inscription", "writing", "stone")
		inscription.Takeable = false
		sanctum.Objects = append(sanctum.Objects, inscription)
	}
	if museum := w.Rooms["modern_museum"]; museum != nil {
		artifact := NewObject("museum_black_glass_mask", "black-glass funerary mask", "The museum label calls it an unprovenanced funerary object. Behind the glass, its polished surface reflects a shoreline instead of the gallery. A faint thumbprint rests on the inside, where no visitor could reach.", "artifact", "mask", "glass mask", "black glass")
		artifact.Takeable = false
		museum.Objects = append(museum.Objects, artifact)
	}

	// Story NPCs live in later regions so exploration reveals lore, not just
	// room descriptions. These figures are guides and witnesses, not required
	// combat encounters.
	charon := NewNPC("charon", "Charon, Ferryman of the Styx", "A narrow boat waits at the bank. The ferryman's face is hidden beneath a hood sewn with tiny silver stars; his pole never touches the riverbed.", 20, 999, 0, 0)
	charon.Faction = "underworld"
	charon.AddDialogue([]string{"hello", "greeting"}, `Charon holds out one pale hand. The nails are clean; the knuckles are scarred, as if he has spent eternity gripping a rope in a storm. "The living ask whether the dead have reached their destination. The dead ask whether anyone noticed they were gone. I can answer neither question. I ferry what arrives, and the river decides what it can keep. If you have come to cross, bring a coin. If you have come to understand, bring a memory you can bear to tell without making yourself the hero of it."`)
	charon.AddDialogue([]string{"fare", "coin", "pay"}, `"The coin is older than the kingdom that minted it," Charon says. "Every age brings me new faces and the same small metal promise. A fare tells the living that a crossing has been acknowledged; it does not buy forgiveness, nor does it bribe the judges. I have carried kings who brought chests of gold and children who had only a button. The river took them both at the same speed. What matters is not what you pay me. It is what you have convinced yourself payment can settle."`)
	charon.AddDialogue([]string{"memory", "remember", "rebirth"}, `The ferryman taps the black thread at your wrist, and the boat shifts although the water is still. "The river takes the shape of what you surrender. I have watched a mother forget the sound of her child's laugh and spend a century trying to remember why she weeps. I have watched a murderer keep every detail of his crime and forget the face of his victim. Memory is not virtue. It is a witness, and witnesses can be frightened, bought, or mistaken. Keep one true memory—not the memory that makes you innocent, but the one that makes you responsible. It may become a shore when all others vanish."`)
	charon.AddDialogue([]string{"cerberus", "hound", "tartarus"}, `"Three heads guard one gate, but each listens for a different fear. The first hears the fear you confess. The second hears the fear you disguise as anger. The third hears the fear you have mistaken for courage. People bring blades to that gate because a blade is easier to understand than a question. Remember this: Cerberus is not the jailer of the dead. He is the promise that what has been buried will not return without being faced."`)
	charon.AddDialogue([]string{"fates", "moirai", "thread"}, `Charon looks toward the far bank. "Clotho spins, Lachesis measures, Atropos cuts—or so the living tell it, because a simple story is easier to survive than the truth. I have ferried souls who swore the Fates had chosen every step, and souls who swore they had chosen every step themselves. Both arrived at my shore. I do not know whether the thread commands the hand or the hand changes the thread. I know only that every choice leaves something for the river to carry."`)
	w.NPCs["styx_shore"] = append(w.NPCs["styx_shore"], charon)

	shade := NewNPC("asphodel_shade", "Unremembered Shade", "A translucent figure kneels among pale flowers, turning an old clay token over and over as if its markings might restore a lost name.", 8, 100, 0, 0)
	shade.Faction = "underworld"
	shade.AddDialogue([]string{"hello", "greeting"}, `The shade looks up from the clay token. "I remember a doorway painted red. I remember barley bread cooling beneath a cloth. I remember someone singing off-key because they thought I was afraid of the dark. I cannot remember the singer's face, or whether the voice belonged to a sister, a father, or a friend. The judges asked me what I loved. I gave them three fragments and watched them write down nothing. Tell me, traveler: if the world forgets the name of a person, does the person become less real—or does the world become less honest?"`)
	shade.AddDialogue([]string{"name", "memory", "remember"}, `"I tried to make a list," the shade says. "A mother's hands. The smell of rain on a roof. The rough edge of a cup repaired with lead. A promise I made and broke. The list became a room, and the room became a house, and then the river entered through every door. I do not ask you to restore my name. A name can be given by strangers and stolen by kings. Give me one memory that can survive being told aloud. Not a grand deed. Something ordinary enough that no poet would bother to keep it—and true enough that someone might recognize the life inside it."`)
	shade.AddDialogue([]string{"judgment", "judge", "asphodel"}, `"The judges ask whether we chose well when we could have chosen otherwise. It is a clean question for a dirty life. I chose badly once, perhaps; or perhaps I was frightened, or hungry, or told that there was no other road. I cannot remember which. That is the cruelty of this field: not punishment, but the endless possibility that I have forgotten the one fact that would let me understand myself. If you reach the halls of judgment, ask them whether a verdict means anything when the witness has been emptied."`)
	shade.AddDialogue([]string{"token", "clay", "doorway", "song"}, `The shade turns the token over. Three grooves cross its face, one straight, one broken, one unfinished. "I made this for someone who was leaving. Or someone who had already left. I remember the clay beneath my fingernails and the song I sang to keep my hands steady. I cannot remember whether the token was meant to open a door or remind someone to return. Perhaps that is why I remain here: not because I have no name, but because I cannot tell whether I was saying goodbye or asking someone to stay."`)
	w.NPCs["fields_asphodel"] = append(w.NPCs["fields_asphodel"], shade)

	curator := NewNPC("museum_curator", "Dr. Eleni Vassos", "A night curator stands beside a glass case whose bronze contents fog the inside of the glass despite the cold gallery.", 5, 100, 0, 0)
	curator.Faction = "mortals"
	curator.AddDialogue([]string{"hello", "greeting"}, `"The museum is closed," Eleni says, still watching the case. "The official reason is a humidity fault. The real reason is that three nights ago the mirror reflected a woman standing behind me. I turned around; the gallery was empty. The next morning the security footage showed me speaking to someone for eleven minutes. I have reviewed it frame by frame. My lips move. The room is empty. You have the expression of someone an artifact has been waiting for, and I have reached the point where I prefer impossible explanations to dishonest ones."`)
	curator.AddDialogue([]string{"artifact", "bronze", "relic"}, `"The catalogue calls it a votive mirror from the fifth century BCE. The bronze is older than the inscription, and the inscription is older than the alphabet used to describe it. It was found beneath a collapsed shrine with no dedication and no human remains. The mirror calls itself a door. I stopped correcting it after the reflection began moving first. I think it was made to remember a threshold, not to show a face. When you stand near it, do not ask what you look like. Ask what the mirror expects you to become."`)
	curator.AddDialogue([]string{"thread", "fate", "past"}, `Eleni opens a drawer and shows you photographs of four unrelated objects: a funerary stele, a Byzantine icon, a nineteenth-century portrait, and a phone camera image taken in the metro. In each reflection, a figure wears a dark thread around the wrist. "The dates are genuine. The materials are genuine. The people who donated them had no connection. The official explanation is restoration error, which is academic language for 'we would rather be wrong in a familiar way.' I am beginning to suspect the thread does not travel through time. I suspect it makes time notice the person wearing it."`)
	curator.AddDialogue([]string{"metro", "tunnel", "styx"}, `"There is a service stair beneath the old station. It is not on any city plan. Every camera pointed at it records the same frame: black water and a single coin. I went down once with a maintenance engineer. We counted forty-two steps. Coming back, we counted forty-three. The engineer resigned the next morning and left the country. I stayed because I found a passenger manifest on my desk, dated tomorrow, with my name listed among the dead. If you find the stair, do not follow the water merely because it seems to know you. Familiarity is not the same thing as safety."`)
	w.NPCs["modern_museum"] = append(w.NPCs["modern_museum"], curator)

	lunarOracle := NewNPC("lunar_oracle", "The Oracle Beneath the Moon", "A blind lunarOracle sits inside a sanctuary carved into pale lunar stone. Constellations drift across the ceiling although no sky lies above it.", 30, 999, 0, 0)
	lunarOracle.Faction = "olympians"
	lunarOracle.AddDialogue([]string{"hello", "greeting"}, `The lunar lunarOracle turns her face toward a sound that has not yet happened. "You have arrived before your own beginning. The stars cannot agree on your name because they are not looking at the same sky. One remembers the world before the gods divided it; another sees a future in which mortals have forgotten the gods entirely. Both are true somewhere. Neither is complete. Tell me what you carried through the dark, and I may tell you which sky has begun to notice you."`)
	lunarOracle.AddDialogue([]string{"future", "fate", "rebirth"}, `"Each life is not a rung on a ladder. It is a world that could have been. Carry your choices forward, and the worlds will begin to answer one another."`)
	lunarOracle.AddDialogue([]string{"moon", "lunar", "prophecy"}, `The lunarOracle traces a constellation that resembles your black thread. "When the last dawn comes, the sea will cover the oldest road. Find the shore that remembers the first choice."`)
	lunarOracle.AddDialogue([]string{"past", "memory", "lives"}, `"Your former selves are not ghosts to command. They are witnesses. One may have loved a god; another may have killed in a god's name; another may have refused both and paid a price no song recorded. Do not gather your lives as trophies. Each one was a person who woke believing the day belonged to them. If you listen honestly, you will discover that your soul is not a single truth repeated. It is an argument that has not finished."`)
	lunarOracle.AddDialogue([]string{"dawn", "shore", "first choice"}, `The lunarOracle draws a line in the lunar dust and leaves its end open. "When the last dawn comes, the sea will cover the oldest road. Find the shore that remembers the first choice. There you must decide whether your lives are meant to repair one another—or whether the world has been using your rebirths to keep an ancient wound open."`)
	w.NPCs["future_moon"] = append(w.NPCs["future_moon"], lunarOracle)

	return w
}

func (w *World) GetRoom(id string) *Room { return w.Rooms[id] }
