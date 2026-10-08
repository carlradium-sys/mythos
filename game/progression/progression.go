package progression

func XPForLevel(level int) int {
	if level <= 1 {
		return 0
	}
	return 100 * (level - 1) * (level - 1)
}

func LevelFromXP(xp int) int {
	level := 1
	for XPForLevel(level+1) <= xp {
		level++
	}
	return level
}
