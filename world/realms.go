package world

func expandWorld(r map[string]*Room) {
	add:=func(id,name,desc string){r[id]=NewRoom(id,name,desc)}
	link:=func(a,d,b string){r[a].Exits[d]=b}

	add("olympus_road","The Celestial Road","White stone rises above the clouds. Golden footprints lead toward halls no mortal map can contain.")
	add("olympus_hall","Hall of the Olympians","Living marble columns surround an impossible court. The thrones seem to change when you look away.")
	add("olympus_garden","Garden of Ambrosia","Silver-leafed trees grow beside luminous fountains. One mortal flower grows where none should.")
	add("delphi_sanctum","The Delphi Sanctum","The Oracle's tripod is cold. Someone has been waiting here for you.")
	add("temple_dawn","Temple of the First Dawn","A forgotten temple stands above the black river. Its altar bears the same mark as the thread bound to your soul.")
	add("ancient_athens","Ancient Athens","A living city surrounds the Acropolis. Rumors spread about a stranger marked by a black thread.")
	add("styx_shore","The Shores of the Styx","Black water moves without wind. A distant boatman waits beneath a sky without stars.")
	add("underworld_crossroads","The Underworld Crossroads","Three roads divide among the dead, judgment, and a gate marked by enormous claw scratches.")
	add("fields_asphodel","Fields of Asphodel","Pale flowers cover an endless field. Shades wander carrying memories that do not belong to them.")
	add("hall_judgment","Hall of Judgment","Three ancient judges sit beneath scales that weigh more than deeds.")
	add("cerberus_gate","Gate of Cerberus","A colossal gate blocks the deepest road. Three sets of eyes open. One head growls your name.")
	add("tartarus_edge","Edge of Tartarus","The abyss below is not empty. Something enormous moves behind the walls of the pit.")
	add("modern_plaka","Plaka After Midnight","Tourists fill the streets by day. At night, old families lock their doors when the wrong stars appear.")
	add("modern_acropolis","The Hidden Acropolis","Beneath the tourist paths is a sealed stairway used by people who know Olympus never truly left.")
	add("modern_metro","The Midnight Metro","A train passes without a driver. Its map contains stations that do not exist.")
	add("modern_museum","Museum of Lost Antiquities","Artifacts officially called replicas are warm to the touch.")
	add("modern_rooftop","The Stormline Rooftops","Athens stretches beneath you. One impossible constellation follows your movements.")
	add("modern_styx","The Modern Styx","A forgotten tunnel descends into black water. Ancient coins sit beside discarded phones.")
	add("modern_sanctum","The Hidden Pantheon","A concrete chamber conceals an older temple. Its occupants have waited for reincarnated souls for generations.")
	add("future_city","Athens, After the Last Dawn","Glass towers rise above temples protected like sacred machines. Humanity has learned that the gods are real.")
	add("future_skyway","The Skyway of Olympus","Humanity's road to Olympus hangs above the clouds. Technology and divine architecture are indistinguishable.")
	add("future_moon","The Lunar Oracle","A sanctuary beneath the moon holds prophecies older than Earth. Your previous lives appear in its walls.")
	add("far_era","The Last Shore","The sea has swallowed the old world. A final Greek shrine survives above the waves.")

	link("olympus_gates","up","olympus_road");link("olympus_road","down","olympus_gates");link("olympus_road","up","olympus_hall");link("olympus_hall","down","olympus_road");link("olympus_hall","east","olympus_garden");link("olympus_garden","west","olympus_hall")
	link("oracle_path","north","delphi_sanctum");link("delphi_sanctum","south","oracle_path");link("delphi_sanctum","east","temple_dawn");link("temple_dawn","west","delphi_sanctum")
	link("olympus_gates","east","ancient_athens");link("ancient_athens","west","olympus_gates")
	link("temple_dawn","down","styx_shore");link("styx_shore","up","temple_dawn");link("styx_shore","east","underworld_crossroads");link("underworld_crossroads","west","styx_shore")
	link("underworld_crossroads","north","fields_asphodel");link("fields_asphodel","south","underworld_crossroads");link("underworld_crossroads","east","hall_judgment");link("hall_judgment","west","underworld_crossroads");link("underworld_crossroads","south","cerberus_gate");link("cerberus_gate","north","underworld_crossroads");link("cerberus_gate","down","tartarus_edge");link("tartarus_edge","up","cerberus_gate")
	link("modern_crossroads","north","modern_plaka");link("modern_plaka","south","modern_crossroads");link("modern_plaka","up","modern_acropolis");link("modern_acropolis","down","modern_plaka");link("modern_plaka","east","modern_metro");link("modern_metro","west","modern_plaka");link("modern_metro","east","modern_styx");link("modern_styx","west","modern_metro");link("modern_plaka","west","modern_museum");link("modern_museum","east","modern_plaka");link("modern_acropolis","up","modern_rooftop");link("modern_rooftop","down","modern_acropolis");link("modern_acropolis","in","modern_sanctum");link("modern_sanctum","out","modern_acropolis")
	link("future_city","up","future_skyway");link("future_skyway","down","future_city");link("future_skyway","east","future_moon");link("future_moon","west","future_skyway");link("future_moon","north","far_era");link("far_era","south","future_moon")
}