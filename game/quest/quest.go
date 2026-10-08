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
 RequiredFlag string
}

var all=[]Quest{
 {ID:"black_thread",Name:"The Black Thread",Description:"Something beyond Olympus recognizes the thread bound to your sword.",Goal:"Defeat a creature in the foothills.",TargetRoom:"olympus_foothills",TargetEnemy:"Harpy",Required:1,RewardXP:80},
 {ID:"oracle_whisper",Name:"The Oracle's Whisper",Description:"The path to Delphi carries a prophecy meant for a soul that has already died.",Goal:"Reach the Delphi Sanctum.",TargetRoom:"delphi_sanctum",Required:1,RewardXP:120},
 {ID:"river_of_memory",Name:"The River of Memory",Description:"The Styx remembers every life you have not yet lived.",Goal:"Reach the Shores of the Styx.",TargetRoom:"styx_shore",Required:1,RewardXP:200},
 {ID:"oath_across_the_river",Name:"An Oath Across the River",Description:"Pythia entrusted you with a warning. Carry it to the Temple of the First Dawn.",Goal:"Reach the Temple of the First Dawn.",TargetRoom:"temple_dawn",Required:1,RewardXP:180,RequiredFlag:"oracle_trust"},
 {ID:"unwritten_path",Name:"The Unwritten Path",Description:"You rejected prophecy. Prove that your own choices can shape the road ahead.",Goal:"Reach Ancient Athens without the Oracle’s blessing.",TargetRoom:"ancient_athens",Required:1,RewardXP:180,RequiredFlag:"oracle_defied"},
}

func All() []Quest { return append([]Quest(nil),all...) }
func Get(id string)(Quest,bool){for _,q:=range all{if q.ID==id{return q,true}};return Quest{},false}
