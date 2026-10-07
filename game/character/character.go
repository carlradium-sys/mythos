package character

type Character struct {
	Name   string
	Level  int
	HP     int
	MaxHP  int
	RoomID string
}

func New(name string) *Character {
	return &Character{
		Name:   name,
		Level:  1,
		HP:     100,
		MaxHP:  100,
		RoomID: "olympus_gates",
	}
}
