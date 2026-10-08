package world
import "fmt"
func(w *World)MapText(cur string)string{m:=map[string]string{"olympus_gates":"[G]","olympus_foothills":"[F]","oracle_path":"[O]","manticore_den":"[M]","modern_crossroads":"[C]"};out:="";for _,id:=range[]string{"oracle_path","olympus_foothills","olympus_gates","manticore_den","modern_crossroads"}{x:=m[id];if id==cur{x="\x1b[1;36m"+x+"\x1b[0m"};out+=fmt.Sprintf("%s ",x)};return out+"\r\nLegend: G Gates F Foothills O Oracle M Manticore C Crossroads"}
