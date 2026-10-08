package character

// Character represents the persistent identity of a player across incarnations.
type Character struct {
	Name       string
	Level      int
	HP         int
	MaxHP      int
	RoomID     string
	Life       int
	Rebirths   int
	Era        string
	Experience int
}

func New(name string) *Character {
	return &Character{
		Name:   name,
		Level:  1,
		HP:     100,
		MaxHP:  100,
		RoomID: "olympus_gates",
		Life:    1,
		Era:     "ancient",
	}
}

func (c *Character) CanRebirth() bool {
	return c.Level >= 10
}

func (c *Character) Rebirth() {
	c.Life++
	c.Rebirths++
	c.Level = 1
	c.Experience = 0
	c.HP = c.MaxHP
	if c.Life == 2 {
		c.Era = "modern"
	}
}
