package session

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"fatewalker/game/character"
	"fatewalker/game/combat"
	"fatewalker/game/item"
	"fatewalker/game/power"
	"fatewalker/game/progression"
	"fatewalker/world"
)

type Session struct {
	Character *character.Character
	Conn net.Conn
	World *world.World
	Enemy *combat.Enemy
}

func New(c *character.Character, conn net.Conn, w *world.World) *Session {
	return &Session{Character:c, Conn:conn, World:w}
}

func (s *Session) WriteLine(format string,args ...any) { fmt.Fprintf(s.Conn,format+"\r\n",args...) }

func (s *Session) Run(scanner *bufio.Scanner) {
	s.WriteLine("")
	s.WriteLine("\x1b[1;35m========================================\x1b[0m")
	s.WriteLine("\x1b[1;36m        FATEWALKER: BEYOND THE STYX\x1b[0m")
	s.WriteLine("\x1b[1;35m========================================\x1b[0m")
	s.WriteLine("")
	s.WriteLine("Welcome, mortal.")
	s.WriteLine("Your first life begins in Ancient Greece.")
	s.WriteLine("Somewhere beyond death, the Fates are already writing your next name.")
	s.WriteLine("")
	s.WriteLine("Type 'help' for commands.")
	s.WriteLine("What is your name?")
	if !scanner.Scan(){return}
	if name:=strings.TrimSpace(scanner.Text()); name!="" { s.Character.Name=name }
	s.look()
	for {
		s.WriteLine("")
		s.WriteLine("%s",s.prompt())
		if !scanner.Scan(){return}
		if !s.handleCommand(strings.TrimSpace(scanner.Text())) {return}
	}
}

func (s *Session) prompt() string {
	return fmt.Sprintf("\x1b[1;36m[%s L%d Life %d HP %d/%d]\x1b[0m > ",s.Character.Era,s.Character.Level,s.Character.Life,s.Character.HP,s.Character.MaxHP)
}

func (s *Session) handleCommand(input string) bool {
	parts:=strings.Fields(strings.ToLower(input))
	if len(parts)==0{return true}
	switch parts[0] {
	case "quit","exit": s.WriteLine("Farewell, %s.",s.Character.Name); return false
	case "help","?": s.help()
	case "look","l": s.look()
	case "map": s.WriteLine(s.World.MapText(s.Character.RoomID))
	case "north","south","east","west","up","down","n","s","e","w","u","d": s.move(parts[0])
	case "who": s.WriteLine("You are the first known traveler on this shard.")
	case "score","stats": s.score()
	case "inventory","i": s.inventory()
	case "attack","kill","hit": s.attack()
	case "cast": s.cast(parts[1:])
	case "powers","power": s.powers()
	case "awaken": s.awaken(parts[1:])
	case "flee": s.flee()
	case "rebirth": s.rebirth()
	default: s.WriteLine("Unknown command. Type 'help' for help.")
	}
	return true
}

func (s *Session) help() {
	s.WriteLine("\x1b[1;33mCommands\x1b[0m")
	s.WriteLine("  look/l          Examine your surroundings")
	s.WriteLine("  map             Show the living world map")
	s.WriteLine("  north/south...  Travel")
	s.WriteLine("  attack          Attack a nearby enemy")
	s.WriteLine("  cast <power>    Invoke an awakened divine power")
	s.WriteLine("  powers          List powers and unlocks")
	s.WriteLine("  awaken <domain> Choose Storm, Tide, Ember, or Aegis at level 3")
	s.WriteLine("  flee            Leave combat")
	s.WriteLine("  inventory/i     Show your gear")
	s.WriteLine("  score/stats     Character and divinity")
	s.WriteLine("  rebirth         Cross the Styx when eligible")
	s.WriteLine("  quit            Leave Fatewalker")
}

func (s *Session) look() {
	r:=s.World.GetRoom(s.Character.RoomID)
	if r==nil {s.WriteLine("You are nowhere. The world has lost track of you.");return}
	s.WriteLine("\x1b[1;33m%s\x1b[0m",r.Name)
	s.WriteLine("%s",r.Description)
	exits:=make([]string,0,len(r.Exits))
	for direction:=range r.Exits {exits=append(exits,direction)}
	if len(exits)>0{s.WriteLine("Exits: %s",strings.Join(exits,", "))}
	if s.Character.RoomID=="manticore_den" && s.Enemy==nil {
		s.Enemy=combat.NewManticore()
		s.WriteLine("\x1b[1;31mA MANTICORE EMERGES FROM THE SHADOWS!\x1b[0m")
		s.WriteLine("%s",s.Enemy.Description)
	}
}

func (s *Session) move(direction string) {
	aliases:=map[string]string{"n":"north","s":"south","e":"east","w":"west","u":"up","d":"down"}
	if v,ok:=aliases[direction];ok{direction=v}
	r:=s.World.GetRoom(s.Character.RoomID)
	if r==nil{return}
	next,ok:=r.Exits[direction]
	if !ok {s.WriteLine("You cannot go that way.");return}
	if s.Enemy!=nil && s.Enemy.HP>0 {s.WriteLine("You cannot leave while the %s still stands.",s.Enemy.Name);return}
	s.Character.RoomID=next
	s.look()
}

func (s *Session) attack() {
	if s.Enemy==nil || s.Enemy.HP<=0 {s.WriteLine("There is nothing here to fight.");return}
	w:=s.currentWeapon()
	result:=combat.Attack(s.Character.Name,w,s.Enemy,s.Character.AttackPower())
	s.WriteLine("%s",result.Text)
	if result.Killed {
		s.Character.AddExperience(s.Enemy.XP)
		s.WriteLine("\x1b[1;32mVictory! +%d XP.\x1b[0m",s.Enemy.XP)
		if s.Character.Level>=5 && len(s.Character.Inventory)==1 {
			drop:=item.Item{Name:"manticore fang",Tier:item.Rare,Damage:18,Kind:"sword"}
			s.Character.Inventory=append(s.Character.Inventory,drop)
			s.WriteLine("%sYou recover a %s%s.",s.color(drop.TierColor()),drop.TierName(),s.color("0"))
			s.WriteLine("The weapon hums faintly, as if it remembers the creature.")
		}
		return
	}
	s.enemyTurn()
}

func (s *Session) cast(args []string) {
	if s.Enemy==nil || s.Enemy.HP<=0 {s.WriteLine("There is no enemy to target.");return}
	if s.Character.Domain=="" {s.WriteLine("You have no awakened divine domain. Try 'awaken storm'.");return}
	if len(args)==0 {s.WriteLine("Cast which power? Try 'powers'.");return}
	var chosen *power.Power
	for _,p:=range power.Unlocked(s.Character.Domain,s.Character.Level) {
		if strings.EqualFold(p.Name,args[0]) || strings.EqualFold(strings.ReplaceAll(p.Name," ",""),args[0]) {q:=p;chosen=&q;break}
	}
	if chosen==nil {s.WriteLine("That power is not yet awakened.");return}
	if s.Character.Mana<chosen.ManaCost {s.WriteLine("Your divine reserves are exhausted.");return}
	s.Character.Mana-=chosen.ManaCost
	result:=combat.Cast(s.Character.Name,chosen.Name,chosen.Damage,string(chosen.Domain),s.Enemy)
	s.WriteLine("\x1b[1;35m%s\x1b[0m",result.Text)
	if result.Killed {
		s.Character.AddExperience(s.Enemy.XP)
		s.WriteLine("\x1b[1;32mDivine victory! +%d XP.\x1b[0m",s.Enemy.XP)
		return
	}
	s.enemyTurn()
}

func (s *Session) enemyTurn() {
	if s.Enemy==nil || s.Enemy.HP<=0{return}
	result:=combat.EnemyAttack(s.Enemy,s.Character.DefensePower())
	if result.Damage>0 {
		s.Character.HP-=result.Damage
		if s.Character.HP<0{s.Character.HP=0}
	}
	s.WriteLine("%s",result.Text)
	if s.Character.HP==0 {
		s.WriteLine("\x1b[1;31mYour mortal life ends here. The Styx waits.\x1b[0m")
		s.Character.Restore()
		s.Character.RoomID="olympus_gates"
		s.Enemy=nil
		s.WriteLine("You awaken at the Gates of Olympus, restored but shaken.")
	}
}

func (s *Session) flee() {
	if s.Enemy==nil || s.Enemy.HP<=0 {s.WriteLine("You are not in combat.");return}
	s.Enemy=nil
	s.WriteLine("You break away from the battle and retreat.")
}

func (s *Session) currentWeapon() item.Item {
	if len(s.Character.Inventory)==0{return item.Item{Name:"fists",Tier:item.Common,Damage:3,Kind:"fist"}}
	for _,w:=range s.Character.Inventory {
		if strings.EqualFold(w.Name,s.Character.Weapon){return w}
	}
	return s.Character.Inventory[0]
}

func (s *Session) inventory() {
	if len(s.Character.Inventory)==0{s.WriteLine("Your inventory is empty.");return}
	s.WriteLine("\x1b[1;33mInventory\x1b[0m")
	for _,i:=range s.Character.Inventory {
		s.WriteLine("%s%s %s(+%d)%s",s.color(i.TierColor()),i.TierName(),i.Name,i.Damage,s.color("0"))
	}
}

func (s *Session) powers() {
	if s.Character.Domain=="" {s.WriteLine("Divinity: dormant. Awaken a domain at level 3.");return}
	s.WriteLine("\x1b[1;35m%s Domain — Divinity %d\x1b[0m",s.Character.Domain,s.Character.Divinity)
	for _,p:=range power.ForDomain(s.Character.Domain) {
		status:="LOCKED"
		if s.Character.Level>=p.UnlockLevel {status="UNLOCKED"}
		s.WriteLine("  %s — level %d — %s",p.Name,p.UnlockLevel,status)
	}
}

func (s *Session) awaken(args []string) {
	if s.Character.Level<3 {s.WriteLine("Something stirs within you, but it is too early to name it. Reach level 3.");return}
	if s.Character.Domain!="" {s.WriteLine("Your soul already bears the %s domain.",s.Character.Domain);return}
	if len(args)==0 {s.WriteLine("Choose storm, tide, ember, or aegis.");return}
	domains:=map[string]power.Domain{"storm":power.Storm,"tide":power.Tide,"ember":power.Ember,"aegis":power.Aegis}
	d,ok:=domains[args[0]]
	if !ok {s.WriteLine("The Fates do not recognize that domain.");return}
	s.Character.LearnDomain(d)
	s.WriteLine("\x1b[1;35mSomething ancient answers your blood.\x1b[0m")
	s.WriteLine("You awaken the %s domain.",d)
	s.WriteLine("Your first power will unlock at level 5.")
}

func (s *Session) score() {
	c:=s.Character
	s.WriteLine("\x1b[1;33m%s\x1b[0m",c.Name)
	s.WriteLine("Level: %d   Life: %d   Rebirths: %d",c.Level,c.Life,c.Rebirths)
	s.WriteLine("Era: %s   HP: %d/%d   Mana: %d/%d",c.Era,c.HP,c.MaxHP,c.Mana,c.MaxMana)
	s.WriteLine("XP: %d   Next level: %d",c.Experience,progression.XPToNextLevel(c.Level,c.Experience))
	s.WriteLine("Attack: %d   Defense: %d   Divinity: %d   Domain: %s",c.AttackPower(),c.DefensePower(),c.Divinity,c.Domain)
}

func (s *Session) rebirth() {
	if !s.Character.CanRebirth() {s.WriteLine("The Fates have not yet opened the way. Reach level 10.");return}
	old:=s.Character.Life
	s.Character.Rebirth()
	s.Enemy=nil
	s.WriteLine("\x1b[1;35mThe world falls silent. You cross the Styx.\x1b[0m")
	s.WriteLine("Life %d ends. Life %d begins.",old,s.Character.Life)
	s.WriteLine("You awaken in the %s era.",s.Character.Era)
	if s.Character.Life==2 {
		s.WriteLine("The ancient world is gone. Athens has become a city of glass, engines, and hidden gods.")
	} else {
		s.WriteLine("Death has stopped being a reset. Your earlier lives now echo through the next.")
	}
}

func (s *Session) color(code string) string { return "\x1b["+code+"m" }
