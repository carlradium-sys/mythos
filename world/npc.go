package world

import "fatewalker/game/item"

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
 Dialogues []Dialogue
 Shop []item.Item
}
func NewNPC(id,name,description string,level,hp,attack,xp int)*NPC{return &NPC{ID:id,Name:name,Description:description,Level:level,HP:hp,MaxHP:hp,Attack:attack,XP:xp}}
func(n *NPC) AddDialogue(keywords []string,text string){n.Dialogues=append(n.Dialogues,Dialogue{Keywords:keywords,Text:text})}
func(n *NPC) DialogueFor(topic string)string{for _,d:=range n.Dialogues{for _,k:=range d.Keywords{if k==topic{return d.Text}}};return ""}
func(n *NPC) HasShop()bool{return len(n.Shop)>0}
