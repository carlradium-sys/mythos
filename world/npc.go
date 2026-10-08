package world

import (
	"strings"

	"fatewalker/game/item"
)

type Dialogue struct { Keywords []string; Text string }
type NPC struct {
 ID string
 Name string
 Description string
 Level int
 HP int
 MaxHP int
 Attack int
 XP int
 Faction string
 Dialogues []Dialogue
 Shop []item.Item
}
func NewNPC(id,name,description string,level,hp,attack,xp int)*NPC{return &NPC{ID:id,Name:name,Description:description,Level:level,HP:hp,MaxHP:hp,Attack:attack,XP:xp}}
func(n *NPC) AddDialogue(keywords []string,text string){n.Dialogues=append(n.Dialogues,Dialogue{Keywords:keywords,Text:text})}
func (n *NPC) DialogueFor(topic string) string {
	normalize := func(value string) string {
		fields := strings.Fields(strings.ToLower(value))
		for i := range fields {
			fields[i] = strings.Trim(fields[i], ".,!?;:\"'()[]{}")
		}
		return strings.Join(fields, " ")
	}
	topic = normalize(topic)
	for _, dialogue := range n.Dialogues {
		for _, keyword := range dialogue.Keywords {
			keyword = normalize(keyword)
			if keyword == "" {
				continue
			}
			if topic == keyword || strings.Contains(" "+topic+" ", " "+keyword+" ") {
				return dialogue.Text
			}
		}
	}
	return ""
}
func(n *NPC) HasShop()bool{return len(n.Shop)>0}
