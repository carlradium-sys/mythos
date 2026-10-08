package world

import "fmt"

func (w *World) MapText(cur string) string {
	rooms := map[string]string{
		"oracle_path":"[O]", "olympus_foothills":"[F]", "olympus_gates":"[G]",
		"manticore_den":"[M]", "modern_crossroads":"[C]",
	}
	out := ""
	for _, id := range []string{"oracle_path","olympus_foothills","olympus_gates","manticore_den","modern_crossroads"} {
		mark:=rooms[id]
		if id==cur { mark="[1;36m"+mark+"[0m" }
		out+=fmt.Sprintf("%s ",mark)
	}
	return out+"\r\nLegend: G Gates | F Foothills | O Oracle | M Manticore | C Crossroads"
}
