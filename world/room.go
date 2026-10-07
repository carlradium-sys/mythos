package world

type Room struct {
	ID          string
	Name        string
	Description string
	Exits       map[string]string
}

func NewRoom(id, name, description string) *Room {
	return &Room{
		ID:          id,
		Name:        name,
		Description: description,
		Exits:       make(map[string]string),
	}
}
