package quest

type Quest struct {
 ID string
 Name string
 Description string
 Goal string
 TargetRoom string
 TargetEnemy string
 Required int
 RewardXP int
 RewardGold int
 RequiredFlag string
 RequiredQuest string
 CompletionFlag string
}

var all=[]Quest{
 {ID:"coin_in_the_ash",Name:"A Coin in the Ash",Description:"A scratched bronze coin lay half-buried in the remains of a cold campfire. Its warmth makes no sense, and its damaged face seems deliberate.",Goal:"Take the coin from the ash, then examine its scratched mark and compare it with the thread around your wrist.",Required:1,RewardXP:35,RewardGold:12,RequiredFlag:"found_scratched_coin",CompletionFlag:"examined_scratched_coin"},
 {ID:"black_thread",Name:"The Black Thread",Description:"Asterion's old well reflected a shore you have never visited, and the Harpy guards something in the ash rather than attacking without warning. The black thread at your wrist seems to connect these impossible details.",Goal:"Decide how to deal with the Harpy, then investigate what its strange habits and the foothills ash may reveal.",TargetRoom:"olympus_foothills",TargetEnemy:"Harpy",Required:1,RewardXP:80,RewardGold:10},
 {ID:"oracle_whisper",Name:"The Oracle's Whisper",Description:"Pythia recognizes a thread that has crossed the river before. Her warning suggests that the gods' gifts and the Fates' consequences are not the same thing.",Goal:"Reach the Delphi Sanctum and learn what the Oracle can—and cannot—tell you about the thread.",TargetRoom:"delphi_sanctum",Required:1,RewardXP:120,RewardGold:15},
 {ID:"river_of_memory",Name:"The River of Memory",Description:"Pythia warned that the Styx can take a memory without taking the pain attached to it. Something about your soul has already survived the crossing.",Goal:"Reach the Shores of the Styx and confront the question of what a soul should be willing to surrender.",TargetRoom:"styx_shore",Required:1,RewardXP:200,RewardGold:25,RequiredQuest:"oracle_whisper"},
 {ID:"oath_across_the_river",Name:"An Oath Across the River",Description:"Pythia entrusted you with a warning. Carry it to the Temple of the First Dawn.",Goal:"Reach the Temple of the First Dawn.",TargetRoom:"temple_dawn",Required:1,RewardXP:180,RewardGold:22,RequiredFlag:"oracle_trust"},
 {ID:"unwritten_path",Name:"The Unwritten Path",Description:"You rejected prophecy. Prove that your own choices can shape the road ahead.",Goal:"Reach Ancient Athens without the Oracle’s blessing.",TargetRoom:"ancient_athens",Required:1,RewardXP:180,RewardGold:22,RequiredFlag:"oracle_defied"},
 {ID:"echoes_in_glass",Name:"Echoes in Glass",Description:"The Styx has left a memory in your soul. Find where the modern river hides beneath Athens.",Goal:"Find the modern Styx in Athens after rebirth.",TargetRoom:"modern_styx",Required:1,RewardXP:260,RewardGold:32,RequiredFlag:"styx_memory_recovered"},
 {ID:"museum_echo",Name:"Museum Echo",Description:"An old object seems to recognize you.",Goal:"Examine the black-glass funerary mask and discover why its reflection shows a shore.",Required:1,RewardXP:180,RewardGold:22,RequiredFlag:"styx_memory_recovered",CompletionFlag:"examined_museum_artifact"},
 {ID:"delphi_trials",Name:"Trials of the Seer",Description:"Pythia tests whether you can hear the truth beneath a prophecy.",Goal:"Examine the broken inscription and listen to the silence between its surviving lines.",Required:1,RewardXP:90,RewardGold:11,RequiredFlag:"room_discovered_delphi_sanctum",CompletionFlag:"examined_delphi_inscription"},
 {ID:"gate_of_three",Name:"The Gate of Three",Description:"Defeat the guardian beneath the river.",Goal:"Defeat Cerberus.",TargetRoom:"cerberus_gate",TargetEnemy:"Cerberus",Required:1,RewardXP:400,RewardGold:50,RequiredFlag:"styx_memory_recovered"},
 {ID:"forge_remembers",Name:"The Forge Remembers",Description:"Theron believes the monsters of Olympus leave behind materials that can be shaped into protection.",Goal:"Craft one piece of equipment at Theron's Forge.",Required:1,RewardXP:140,RewardGold:17,RequiredFlag:"forge_discovered",CompletionFlag:"crafted_any_gear"},
 {ID:"charons_memory",Name:"A Memory for the Dead",Description:"Charon believes a true memory can become a shore for a soul eroded by the Styx. In Asphodel, one shade has kept fragments of an ordinary life but cannot tell whether they are enough to prove that life mattered.",Goal:"Hear Charon's account of memory and responsibility, then carry its meaning to the Unremembered Shade in Asphodel.",Required:1,RewardXP:160,RewardGold:20,RequiredFlag:"charon_memory_heard",CompletionFlag:"asphodel_memory_returned"},
 {ID:"return_the_coin",Name:"An Offering Returned",Description:"The scratched coin has the shape of a ferry fare, yet it was found beside a village well that reflects a shore no one remembers. Damon says the village leaves bread and salt at the well for the missing. You can return the coin to the offering bowl and see what answers, or keep it and follow the mystery elsewhere.",Goal:"Carry the examined scratched coin back to Asterion's square and offer it to the clay bowl beside the old well.",Required:1,RewardXP:65,RewardGold:8,RequiredFlag:"examined_scratched_coin",CompletionFlag:"offered_coin_at_well"},
}

func All() []Quest { return append([]Quest(nil),all...) }
func Get(id string)(Quest,bool){for _,q:=range all{if q.ID==id{return q,true}};return Quest{},false}
