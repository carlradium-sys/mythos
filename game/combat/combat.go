package combat

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"fatewalker/game/item"
)

type Enemy struct {
	Name, Description string
	Level, HP, MaxHP, Defense, Damage, XP int
	CanSever, TailSevered bool
}

type Result struct {
	Damage int
	Text string
	Killed bool
}

var (
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	rngMu sync.Mutex
)

func roll(n int) int {
	rngMu.Lock()
	defer rngMu.Unlock()
	return rng.Intn(n)
}

func NewHarpy() *Enemy { return &Enemy{Name:"Harpy", Description:"A storm-winged creature circles the ruins, watching for an opening.", Level:2, HP:75, MaxHP:75, Defense:5, Damage:10, XP:70} }

func NewSatyr() *Enemy { return &Enemy{Name:"Satyr", Description:"A horned guardian blocks the path, staff in hand and eyes bright with old magic.", Level:3, HP:110, MaxHP:110, Defense:6, Damage:13, XP:110} }

func NewManticore() *Enemy {
	return &Enemy{Name:"Manticore", Description:"A lion-bodied horror watches you with a humanlike face. A crown of venomous spines ripples along its tail.", Level:4, HP:170, MaxHP:170, Defense:7, Damage:18, XP:180, CanSever:true}
}

func Attack(name string, w item.Item, e *Enemy, attack int) Result {
	if e.HP <= 0 { return Result{Text:"The "+e.Name+" is already defeated."} }
	if roll(20)+1+attack < e.Defense+10 { return Result{Text:fmt.Sprintf("Your %s misses; the %s twists away from the blow. (0 damage)", w.Name, e.Name)} }
	damage := w.Damage + roll(7)
	critical := roll(100) < 12
	if critical { damage += damage/2 }
	locations := []string{"wing","side","shoulder","foreleg","rib cage","neck"}
	if e.CanSever && !e.TailSevered { locations=append(locations,"tail") }
	where := locations[roll(len(locations))]
	e.HP -= damage
	if e.HP < 0 { e.HP=0 }
	verb := "strikes"
	if w.Kind=="bow" { verb="pierces" } else if w.Kind=="sword" { verb="slashes" } else if w.Kind=="spear" { verb="drives into" }
	text := fmt.Sprintf("Your %s %s the %s's %s.", w.Name,verb,e.Name,where)
	if critical { text+=" The impact is devastating." }
	text += fmt.Sprintf(" (%d damage)", damage)
	if where=="tail" && roll(100)<18 { e.TailSevered=true; text+=" The blade cuts through—the manticore's tail is severed!" }
	if e.HP==0 { text+=" The "+e.Name+" collapses." }
	return Result{Damage:damage,Text:text,Killed:e.HP==0}
}

func Cast(caster,pName string,dmg int,domain string,e *Enemy) Result {
	damage:=dmg+roll(10)
	e.HP-=damage
	if e.HP<0 { e.HP=0 }
	var text string
	switch domain {
	case "Storm": text=fmt.Sprintf("Lightning leaps across the room, lighting the chamber, then passes through the %s and bursts into sparks behind it.",e.Name)
	case "Tide": text=fmt.Sprintf("A roaring wall of water coils from %s's hands and hammers the %s backward.",caster,e.Name)
	case "Ember": text=fmt.Sprintf("A star of living fire crosses the room and strikes the %s in a flash of heat.",e.Name)
	case "Aegis": text=fmt.Sprintf("%s releases a shockwave of divine force that buckles the air around the %s.",caster,e.Name)
	default: text=fmt.Sprintf("%s unleashes %s against the %s.",caster,pName,e.Name)
	}
	text+=fmt.Sprintf(" (%d damage)",damage)
	if e.HP==0 { text+=" The creature falls." }
	return Result{Damage:damage,Text:text,Killed:e.HP==0}
}

func EnemyAttack(e *Enemy, defense int) Result {
	approach, impact := enemyAttackNarration(e.Name)
	if roll(20)+1+e.Level < defense+8 {
		return Result{Text:fmt.Sprintf("%s You evade the attack. (0 damage)", approach)}
	}
	damage := e.Damage/2 + roll(e.Damage/2+1)
	return Result{Damage: damage, Text: fmt.Sprintf("%s (%d damage)", impact, damage)}
}

func enemyAttackNarration(name string) (string, string) {
	switch name {
	case "Harpy":
		return "The Harpy folds its wings and stoops from above.", "Talons rake across you as the Harpy beats past"
	case "Satyr":
		return "The Satyr plants its hooves and whips its rune-carved staff around.", "The staff crashes into you with a burst of green sparks"
	case "Manticore":
		return "The Manticore's tail arches over its shoulder, spines hissing through the air.", "Venomous spines punch into you as the Manticore's tail snaps back"
	default:
		return "The " + name + " lunges for an opening.", "The " + name + " lands a heavy blow"
	}
}
