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
func normalizeDialogueText(value string) string {
	fields := strings.Fields(strings.ToLower(value))
	for i := range fields {
		fields[i] = strings.Trim(fields[i], ".,!?;:\\\"'()[]{}")
	}
	return strings.Join(fields, " ")
}

// MatchDialogue resolves natural-language questions to a known dialogue intent.
// Extra wording is fine when the question includes a meaningful topic keyword.
func (n *NPC) MatchDialogue(topic string) (string, string) {
	topic = normalizeDialogueText(topic)
	if topic == "" { return "", "" }
	for _, dialogue := range n.Dialogues {
		for _, rawKeyword := range dialogue.Keywords {
			keyword := normalizeDialogueText(rawKeyword)
			if keyword == "" { continue }
			if topic == keyword || strings.Contains(" "+topic+" ", " "+keyword+" ") {
				return keyword, dialogue.Text
			}
		}
	}
	return "", ""
}

func (n *NPC) DialogueFor(topic string) string {
	_, text := n.MatchDialogue(topic)
	return text
}
func(n *NPC) HasShop()bool{return len(n.Shop)>0}
