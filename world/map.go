package world

import "fmt"

func (w *World) MapText(cur string) string {
 groups:=[]struct{name string; ids []string}{
  {"ANCIENT",[]string{"olympus_gates","olympus_foothills","oracle_path","delphi_sanctum","manticore_den","temple_dawn","ancient_athens","olympus_road","olympus_hall","olympus_garden","styx_shore","underworld_crossroads","fields_asphodel","hall_judgment","cerberus_gate","tartarus_edge"}},
  {"MODERN",[]string{"modern_crossroads","modern_plaka","modern_acropolis","modern_metro","modern_museum","modern_rooftop","modern_styx","modern_sanctum"}},
  {"LATER LIVES",[]string{"future_city","future_skyway","future_moon","far_era"}},
 }
 out:="\x1b[1;33mFATEWALKER WORLD MAP\x1b[0m\n"
 for _,g:=range groups {
  out+=fmt.Sprintf("\x1b[1;36m%s\x1b[0m ",g.name)
  for _,id:=range g.ids { mark:="[ ]";if id==cur{mark="\x1b[1;36m[@]\x1b[0m"};out+=mark+" " }
  out+="\n"
 }
 out+="@ current | The world grows as your soul crosses eras."
 return out
}
