package progression

func XPForLevel(level int) int {
	if level <= 1 {
		return 0
	}
	return 100 * (level - 1) * (level - 1)
}

func LevelFromXP(xp int) int {
	level := 1
	for level < 50 && xp >= XPForLevel(level+1) {
		level++
	}
	return level
}

func XPToNextLevel(level, xp int) int {
	next := XPForLevel(level+1) - xp
	if next < 0 {
		return 0
	}
	return next
}
