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
	Kind   string
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
