package combat

import (
	"math/rand"
	"time"

	"fatewalker/game/character"
)

type Enemy struct {
	Name   string
	Level  int
	HP     int
	MaxHP  int
	Attack int
	XP     int
}

func NewEnemy(name string, level, hp, attack, xp int) *Enemy {
	return &Enemy{Name: name, Level: level, HP: hp, MaxHP: hp, Attack: attack, XP: xp}
}

func RollDamage(base int) int {
	if base <= 1 {
		return 1
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return base/2 + r.Intn(base-base/2+1)
}

func PlayerAttack(c *character.Character) int {
	return RollDamage(c.AttackPower())
}

func EnemyAttack(e *Enemy) int {
	return RollDamage(e.Attack)
}
