package item

type Tier int

const (
	Common Tier = iota
	Uncommon
	Rare
	Epic
	Legendary
	Mythic
)

type Item struct {
	Name   string
	Tier   Tier
	Damage int
	Armor  int
	Kind   string
	Price  int
}

func (i Item) TierName() string {
	names := []string{"Common", "Uncommon", "Rare", "Epic", "Legendary", "Mythic"}
	if int(i.Tier) >= 0 && int(i.Tier) < len(names) {
		return names[i.Tier]
	}
	return "Unknown"
}

func (i Item) TierColor() string {
	colors := []string{"37", "32", "34", "35", "33", "91"}
	if int(i.Tier) >= 0 && int(i.Tier) < len(colors) {
		return colors[i.Tier]
	}
	return "37"
}

func StarterArmor() Item { return Item{Name:"linen cuirass", Tier:Common, Armor:2, Kind:"armor"} }
func IsWeapon(i Item) bool { return i.Kind=="sword" || i.Kind=="bow" || i.Kind=="spear" || i.Kind=="fist" }
