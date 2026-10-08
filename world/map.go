package world

import "fmt"

func (w *World) MapText(cur string) string {
 groups:=[]struct{name string; ids []string}{
  {"ANCIENT",[]string{"olympus_gates","olympus_foothills","oracle_path","delphi_sanctum","manticore_den","temple_dawn","ancient_athens","olympus_road","olympus_hall","olympus_garden","styx_shore","underworld_crossroads","fields_asphodel","hall_judgment","cerberus_gate","tartarus_edge"}},
  {"MODERN",[]string{"modern_crossroads","modern_plaka","modern_acropolis","modern_metro","modern_museum","modern_rooftop","modern_styx","modern_sanctum"}},
  {"LATER LIVES",[]string{"future_city","future_skyway","future_moon","far_era"}},
 }
 out:="[1;33mFATEWALKER WORLD MAP[0m
"
 for _,g:=range groups {
  out+=fmt.Sprintf("[1;36m%s[0m ",g.name)
  for _,id:=range g.ids { mark:="[ ]";if id==cur{mark="[1;36m[@][0m"};out+=mark+" " }
  out+="
"
 }
 out+="@ current | The world grows as your soul crosses eras."
 return out
}
