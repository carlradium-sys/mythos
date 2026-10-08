package character

import (
	"fatewalker/game/item"
	"fatewalker/game/power"
	"fatewalker/game/progression"
)

type Character struct {
	Name       string
	Level      int
	HP         int
	MaxHP      int
	Mana       int
	MaxMana    int
	RoomID     string
	Life       int
	Rebirths   int
	Era        string
	Experience int
	Gold int
	Reputation map[string]int
	Attack     int
	Defense    int
	Domain     power.Domain
	Divinity   int
	Weapon     string
	Armor      string
	Inventory  []item.Item
	Memories   []string
	Scars      []string
	Oaths      []string
	Favors     []string
	Curses     []string
	Echoes     []string
	Quests     map[string]int
	StoryFlags map[string]bool
}

func New(name string) *Character {
	return &Character{
		Name: name, Level: 1, HP: 100, MaxHP: 100,
		Mana: 30, MaxMana: 30, RoomID: "olympus_gates",
		Life: 1, Era: "ancient", Attack: 12, Defense: 3,
		Weapon: "bronze sword", Armor: "linen cuirass",
		Quests: map[string]int{},
		StoryFlags: map[string]bool{},
		Reputation: map[string]int{"olympians":0,"delphi":0,"underworld":0},
		Gold: 50,
		Inventory: []item.Item{{Name: "bronze sword", Tier: item.Common, Damage: 10, Kind: "sword"}, item.StarterArmor()},
	}
}

// HasStoryFlag reports whether a persistent story choice or discovery has been recorded.
func (c *Character) HasStoryFlag(flag string) bool {
	return c.StoryFlags != nil && c.StoryFlags[flag]
}

// SetStoryFlag records a choice or discovery on the character's persistent soul.
func (c *Character) SetStoryFlag(flag string) {
	if c.StoryFlags == nil {
		c.StoryFlags = map[string]bool{}
	}
	c.StoryFlags[flag] = true
}

func (c *Character) AttackPower() int { return c.Attack + c.Level*2 }
func (c *Character) DefensePower() int { return c.Defense + c.Level + c.ArmorPower() }
func (c *Character) ArmorPower() int { for _,i:=range c.Inventory { if i.Name==c.Armor { return i.Armor } }; return 0 }
func (c *Character) CanRebirth() bool { return c.Level >= 10 }

func (c *Character) AddExperience(amount int) bool {
	if amount <= 0 {
		return false
	}
	old := c.Level
	c.Experience += amount
	c.Level = progression.LevelFromXP(c.Experience)
	if c.Level > old {
		for level := old + 1; level <= c.Level; level++ {
			c.MaxHP += 12
			c.MaxMana += 4
			c.Attack += 2
			c.Defense++
		}
		c.HP, c.Mana = c.MaxHP, c.MaxMana
		return true
	}
	return false
}

func (c *Character) Restore() { c.HP, c.Mana = c.MaxHP, c.MaxMana }

func (c *Character) LearnDomain(d power.Domain) {
	if c.Domain == "" {
		c.Domain = d
		c.Divinity = 1
	}
}

// Rebirth begins a new life while preserving the soul's accumulated divine identity.
func (c *Character) Rebirth() {
	if c.Life == 1 {
		c.RebirthTo("modern", "modern_crossroads")
		return
	}
	c.RebirthTo("beyond", "modern_crossroads")
}

// RebirthTo lets later lives choose an era without changing the core rebirth rules.
func (c *Character) RebirthTo(era, room string) {
	c.Life++
	c.Rebirths++
	c.Level = 1
	c.Experience = 0
	c.HP, c.Mana = c.MaxHP, c.MaxMana
	c.Attack, c.Defense = 12, 3
	c.Era = era
	c.RoomID = room
	c.Divinity++
	c.Memories = append(c.Memories, "You remember crossing the Styx and hearing the Fates whisper your name.")
	c.Echoes = append(c.Echoes, "A black thread survives another death.")

}
