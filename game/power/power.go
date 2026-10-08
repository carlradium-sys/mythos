package power

type Domain string

const (
	Storm Domain = "Storm"
	Tide  Domain = "Tide"
	Ember Domain = "Ember"
	Aegis Domain = "Aegis"
)

type Power struct {
	Name        string
	Domain      Domain
	UnlockLevel int
	ManaCost    int
	Damage      int
	Description string
}

func ForDomain(d Domain) []Power {
	switch d {
	case Storm:
		return []Power{
			{Name: "Stormspark", Domain: d, UnlockLevel: 5, ManaCost: 8, Damage: 30, Description: "A crackling bolt leaps from your hand."},
			{Name: "Thunderfall", Domain: d, UnlockLevel: 8, ManaCost: 16, Damage: 55, Description: "Thunder answers your call and strikes from above."},
			{Name: "Judgment of the Sky", Domain: d, UnlockLevel: 15, ManaCost: 30, Damage: 95, Description: "The heavens split open and a column of lightning descends."},
		}
	case Tide:
		return []Power{
			{Name: "Undertow", Domain: d, UnlockLevel: 5, ManaCost: 8, Damage: 28, Description: "A crushing ribbon of water coils around your foe."},
			{Name: "Leviathan Surge", Domain: d, UnlockLevel: 8, ManaCost: 16, Damage: 52, Description: "A towering wave crashes through the battlefield."},
			{Name: "Sea of the First Deep", Domain: d, UnlockLevel: 15, ManaCost: 30, Damage: 90, Description: "The battlefield becomes a fragment of the primordial sea."},
		}
	case Ember:
		return []Power{
			{Name: "Cinderbrand", Domain: d, UnlockLevel: 5, ManaCost: 8, Damage: 32, Description: "White-hot flame brands the target."},
			{Name: "Promethean Star", Domain: d, UnlockLevel: 8, ManaCost: 16, Damage: 58, Description: "A miniature sun blooms at your fingertips."},
			{Name: "Titanfire", Domain: d, UnlockLevel: 15, ManaCost: 30, Damage: 100, Description: "Ancient fire answers a name older than Olympus."},
		}
	case Aegis:
		return []Power{
			{Name: "Aegis Pulse", Domain: d, UnlockLevel: 5, ManaCost: 8, Damage: 20, Description: "Divine force erupts outward from your shield."},
			{Name: "Olympian Bulwark", Domain: d, UnlockLevel: 8, ManaCost: 16, Damage: 48, Description: "A radiant wall crashes into your enemy."},
			{Name: "Sovereign Ward", Domain: d, UnlockLevel: 15, ManaCost: 30, Damage: 82, Description: "The old laws of divine protection become a weapon."},
		}
	}
	return nil
}

func Unlocked(d Domain, level int) []Power {
	all := ForDomain(d)
	out := make([]Power, 0, len(all))
	for _, p := range all {
		if level >= p.UnlockLevel {
			out = append(out, p)
		}
	}
	return out
}
