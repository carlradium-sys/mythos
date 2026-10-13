package world

type Room struct {
	ID          string
	Name        string
	Description string
	Exits       map[string]string
	Objects     []*Object
}

func NewRoom(id, name, description string) *Room {
	return &Room{
		ID:          id,
		Name:        name,
		Description: description,
		Exits:       make(map[string]string),
		Objects:     []*Object{},
	}
}
