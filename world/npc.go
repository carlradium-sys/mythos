package world

type NPC struct {
	ID          string
	Name        string
	Description string
	Level       int
	HP          int
	MaxHP       int
	Attack      int
	XP          int
}

func NewNPC(id, name, description string, level, hp, attack, xp int) *NPC {
	return &NPC{ID:id, Name:name, Description:description, Level:level, HP:hp, MaxHP:hp, Attack:attack, XP:xp}
}
