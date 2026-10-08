package character

import "fatewalker/game/progression"

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
	Attack     int
	Defense    int
}

func New(name string) *Character {
	return &Character{
		Name: name, Level: 1, HP: 100, MaxHP: 100,
		RoomID: "olympus_gates", Life: 1, Era: "ancient",
		Attack: 12, Defense: 3,
	}
}

func (c *Character) AttackPower() int { return c.Attack + c.Level*2 }
func (c *Character) DefensePower() int { return c.Defense + c.Level }

func (c *Character) CanRebirth() bool { return c.Level >= 10 }

func (c *Character) AddExperience(amount int) bool {
	if amount <= 0 { return false }
	old := c.Level
	c.Experience += amount
	c.Level = progression.LevelFromXP(c.Experience)
	if c.Level > old {
		for level := old + 1; level <= c.Level; level++ {
			c.MaxHP += 12
			c.Attack += 2
			c.Defense++
		}
		c.HP = c.MaxHP
		return true
	}
	return false
}

func (c *Character) Restore() { c.HP = c.MaxHP }

func (c *Character) Rebirth() {
	c.Life++
	c.Rebirths++
	c.Level = 1
	c.Experience = 0
	c.HP = c.MaxHP
	c.Attack = 12
	c.Defense = 3
	c.RoomID = "modern_crossroads"
	if c.Life == 2 {
		c.Era = "modern"
	} else {
		c.Era = "beyond"
	}
}
