package session

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

	"fatewalker/game/character"
	"fatewalker/game/combat"
	"fatewalker/game/help"
	"fatewalker/game/item"
	"fatewalker/game/power"
	"fatewalker/game/progression"
	"fatewalker/game/account"
	"fatewalker/world"
)

type Session struct {
 Character *character.Character
 Account *account.Account
 Accounts *account.Store
 Conn net.Conn
 World *world.World
 Enemy *combat.Enemy
 TutorialStep int
}

func New(accounts *account.Store, conn net.Conn, w *world.World) *Session {
 return &Session{Accounts:accounts, Conn:conn, World:w}
}

func (s *Session) WriteLine(format string,args ...any) { fmt.Fprintf(s.Conn,format+"\r\n",args...) }

func (s *Session) Run(scanner *bufio.Scanner) {
	s.WriteLine("")
	s.WriteLine("\x1b[1;35m========================================\x1b[0m")
	s.WriteLine("\x1b[1;36m        FATEWALKER: BEYOND THE STYX\x1b[0m")
	s.WriteLine("\x1b[1;35m========================================\x1b[0m")
	s.WriteLine("")
	s.WriteLine("\x1b[1;33mTHE THREAD REMEMBERS\x1b[0m")
	s.WriteLine("Your soul is persistent. The server remembers your account, characters, lives, and history.")
	s.WriteLine("")
	if !s.login(scanner) { return }
	if !s.selectCharacter(scanner) { return }
	s.look()
	s.tutorialHint()
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
	case "quit","exit":
		s.WriteLine("Farewell, %s.",s.Character.Name)
		return false
	case "help","?":
		s.help(parts[1:])
	case "look","l":
		s.look()
		s.advanceTutorial(1)
	case "map":
		s.WriteLine(s.World.MapText(s.Character.RoomID))
	case "north","south","east","west","up","down","n","s","e","w","u","d":
		s.move(parts[0])
	case "who":
		s.WriteLine("You are the first known traveler on this shard.")
	case "score","stats":
		s.score()
	case "inventory","i":
		s.inventory()
	case "attack","kill","hit":
		s.attack()
	case "cast":
		s.cast(parts[1:])
	case "powers","power":
		s.powers()
	case "awaken":
		s.awaken(parts[1:])
	case "flee":
		s.flee()
	case "rebirth":
		s.rebirth(parts[1:])
	case "journal","quest","story":
		s.journal()
	case "soul","legacy":
		s.soul()
	default:
		s.WriteLine("Unknown command. Type 'help' for help.")
	}
	s.persist()
	return true
}

func (s *Session) login(scanner *bufio.Scanner) bool {
 for {
  s.WriteLine("")
  s.WriteLine("\x1b[1;33mACCOUNT\x1b[0m")
  s.WriteLine("Type: login <username> <password>")
  s.WriteLine("Or:   register <username> <password>")
  s.WriteLine("Use 'quit' to disconnect.")
  if !scanner.Scan(){return false}
  p:=strings.Fields(scanner.Text())
  if len(p)==1&&strings.EqualFold(p[0],"quit"){return false}
  if len(p)!=3 {s.WriteLine("Please enter one of the two formats above.");continue}
  var err error
  if strings.EqualFold(p[0],"register") {
   err=s.Accounts.Register(p[1],p[2])
   if err==nil {s.Account,sErr:=s.Accounts.Login(p[1],p[2]);_ = sErr; s.Account=s; s.WriteLine("Account created.")}
  } else if strings.EqualFold(p[0],"login") {
   s.Account,err=s.Accounts.Login(p[1],p[2])
  } else {err=fmt.Errorf("unknown account command")}
  if err!=nil{s.WriteLine("Account error: %s",err);continue}
  s.WriteLine("Logged in as %s.",s.Account.Username)
  return true
 }
}

func (s *Session) selectCharacter(scanner *bufio.Scanner) bool {
 for {
  s.WriteLine("")
  s.WriteLine("\x1b[1;33mCHARACTER SELECT\x1b[0m")
  if len(s.Account.Characters)==0 {s.WriteLine("No characters yet. Type: create <name>")}
  for i,c:=range s.Account.Characters {s.WriteLine("  %d) %s — Life %d, Level %d, %s",i+1,c.Name,c.Life,c.Level,c.Era)}
  s.WriteLine("Commands: <number>, create <name>, delete <number>, logout, quit")
  if !scanner.Scan(){return false}
  p:=strings.Fields(scanner.Text());if len(p)==0{continue}
  switch strings.ToLower(p[0]) {
  case "logout": return s.login(scanner)
  case "quit","exit": return false
  case "create":
   if len(p)<2{s.WriteLine("Create which character?");continue}
   rec,err:=s.Account.NewCharacter(strings.Join(p[1:]," "));if err!=nil{s.WriteLine("Cannot create character: %s",err);continue}
   s.Account.Characters=append(s.Account.Characters,*rec);s.persist();s.Character=rec.Character;s.WriteLine("%s created.",s.Character.Name);return true
  case "delete":
   if len(p)!=2{s.WriteLine("Delete which character number?");continue}
   n,err:=strconv.Atoi(p[1]);if err!=nil||n<1||n>len(s.Account.Characters){s.WriteLine("Invalid character number.");continue}
   s.Account.Characters=append(s.Account.Characters[:n-1],s.Account.Characters[n:]...);s.persist();s.WriteLine("Character deleted.")
  default:
   n,err:=strconv.Atoi(p[0]);if err!=nil||n<1||n>len(s.Account.Characters){s.WriteLine("Choose a character number.");continue}
   s.Character=s.Account.Characters[n-1].Character;s.WriteLine("Welcome back, %s.",s.Character.Name);return true
  }
 }
}

func (s *Session) persist() {
 if s.Account==nil||s.Character==nil{return}
 for i:=range s.Account.Characters {if s.Account.Characters[i].Character==s.Character {break}}
 if err:=s.Accounts.Save(s.Account);err!=nil{s.WriteLine("Persistence warning: %s",err)}
}

func (s *Session) help(args []string) {
	topic := "start"
	if len(args)>0 { topic=strings.Join(args," ") }
	if topic=="list" {
		s.WriteLine("\x1b[1;33mHELP TOPICS\x1b[0m")
		s.WriteLine("  %s",strings.Join(help.Names(),", "))
		s.WriteLine("Use: help <topic>")
		return
	}
	t,ok:=help.Get(topic)
	if !ok {
		s.WriteLine("No help file exists for '%s'. Try 'help list'.",topic)
		return
	}
	s.WriteLine("\x1b[1;33m[%s]\x1b[0m",strings.ToUpper(t.Name))
	for _,line:=range strings.Split(t.Text,"\n") { s.WriteLine("%s",line) }
}

func (s *Session) journal() {
	s.WriteLine("\x1b[1;33mFATEWALKER JOURNAL\x1b[0m")
	s.WriteLine("\x1b[1;36mThe First Thread\x1b[0m")
	s.WriteLine("You awakened at the Gates of Olympus with no memory of your death.")
	s.WriteLine("A black thread binds itself to your bronze sword.")
	s.WriteLine("Something beyond the gates knows your name.")
	s.WriteLine("")
	s.WriteLine("\x1b[1;33mCurrent thread\x1b[0m")
	switch {
	case s.Character.Life == 1 && s.Character.Level < 3:
		s.WriteLine("Explore the foothills. Discover why the creatures seem to recognize you.")
	case s.Character.Life == 1 && s.Character.Level < 10:
		s.WriteLine("Grow stronger, follow the Oracle's path, and uncover the meaning of the black thread.")
	case s.Character.Life == 1:
		s.WriteLine("The Styx is near. Decide what kind of soul you will carry into your next life.")
	case s.Character.Life == 2:
		s.WriteLine("Athens is familiar in ways it should not be. Find the place where the ancient world survived.")
	default:
		s.WriteLine("Your earlier lives are becoming a single story. Find the next thread.")
	}
	s.WriteLine("")
	s.WriteLine("The journal offers direction, not a leash. You are free to wander.")
}

func (s *Session) tutorialHint() {
	switch s.TutorialStep {
	case 0:
		s.WriteLine("\x1b[1;36mTutorial:\x1b[0m Type 'look' to study the place where your story begins.")
	case 1:
		s.WriteLine("\x1b[1;36mTutorial:\x1b[0m The black thread pulls north. Try 'north' when you are ready.")
	case 2:
		s.WriteLine("\x1b[1;36mTutorial:\x1b[0m A creature has noticed you. Try 'attack'.")
	case 3:
		s.WriteLine("\x1b[1;36mTutorial:\x1b[0m You survived. Check 'inventory' and 'score', then continue exploring.")
	}
}

func (s *Session) advanceTutorial(step int) {
	if step>s.TutorialStep {
		s.TutorialStep=step
		s.tutorialHint()
	}
}

func (s *Session) look() {
	r:=s.World.GetRoom(s.Character.RoomID)
	if r==nil {s.WriteLine("You are nowhere. The world has lost track of you.");return}
	s.WriteLine("\x1b[1;33m%s\x1b[0m",r.Name)
	s.WriteLine("%s",r.Description)
	exits:=make([]string,0,len(r.Exits))
	for direction:=range r.Exits {exits=append(exits,direction)}
	if len(exits)>0{s.WriteLine("Exits: %s",strings.Join(exits,", "))}
	if s.Enemy==nil { s.spawnEnemy() }
}

func (s *Session) spawnEnemy() {
	switch s.Character.RoomID {
	case "olympus_foothills":
		s.Enemy=combat.NewHarpy()
	case "oracle_path":
		s.Enemy=combat.NewSatyr()
	case "manticore_den":
		s.Enemy=combat.NewManticore()
	default:
		return
	}
	s.WriteLine("\x1b[1;31mA %s appears!\x1b[0m",s.Enemy.Name)
	s.WriteLine("%s",s.Enemy.Description)
	s.advanceTutorial(2)
}

func (s *Session) canEnter(id string) bool {
	c:=s.Character
	if strings.HasPrefix(id,"modern_") && c.Life<2 { s.WriteLine("The world beyond the Styx has not opened to you. This is a road for another life."); return false }
	if strings.HasPrefix(id,"future_") || id=="far_era" { s.WriteLine("That horizon belongs to a later chapter of your soul."); return false }
	if id=="olympus_road" && c.Level<6 { s.WriteLine("The path into Olympus is veiled by divine law. Become stronger before attempting the ascent."); return false }
	if id=="styx_shore" && c.Level<7 { s.WriteLine("The black river calls, but you are not yet strong enough to cross its threshold."); return false }
	if id=="cerberus_gate" && c.Level<9 { s.WriteLine("A presence beyond the gate warns you away. The guardian is not yet your battle."); return false }
	return true
}

func (s *Session) move(direction string) {
	aliases:=map[string]string{"n":"north","s":"south","e":"east","w":"west","u":"up","d":"down"}
	if v,ok:=aliases[direction];ok{direction=v}
	r:=s.World.GetRoom(s.Character.RoomID)
	if r==nil{return}
	next,ok:=r.Exits[direction]
	if !ok {s.WriteLine("You cannot go that way.");return}
	if s.Enemy!=nil && s.Enemy.HP>0 {s.WriteLine("You cannot leave while the %s still stands.",s.Enemy.Name);return}
	if !s.canEnter(next) { return }
	s.Character.RoomID=next
	s.look()
	if s.Character.RoomID=="olympus_foothills" { s.advanceTutorial(2) }
}

func (s *Session) attack() {
	if s.Enemy==nil || s.Enemy.HP<=0 {s.WriteLine("There is nothing here to fight.");return}
	w:=s.currentWeapon()
	result:=combat.Attack(s.Character.Name,w,s.Enemy,s.Character.AttackPower())
	s.WriteLine("%s",result.Text)
	s.advanceTutorial(3)
	if result.Killed {
		s.Character.AddExperience(s.Enemy.XP)
		s.WriteLine("\x1b[1;32mVictory! +%d XP.\x1b[0m",s.Enemy.XP)
		if s.Enemy.Name=="Manticore" && s.Character.Level>=5 && len(s.Character.Inventory)==1 {
			drop:=item.Item{Name:"manticore fang",Tier:item.Rare,Damage:18,Kind:"sword"}
			s.Character.Inventory=append(s.Character.Inventory,drop)
			s.WriteLine("%sYou recover a %s %s%s.",s.color(drop.TierColor()),drop.TierName(),drop.Name,s.color("0"))
			s.WriteLine("The weapon hums faintly, as if it remembers the creature.")
		}
		s.Enemy=nil
		return
	}
	s.enemyTurn()
}

func (s *Session) cast(args []string) {
	if s.Enemy==nil || s.Enemy.HP<=0 {s.WriteLine("There is no enemy to target.");return}
	if s.Character.Domain=="" {s.WriteLine("You have no awakened divine domain. Try 'awaken storm'.");return}
	if len(args)==0 {s.WriteLine("Cast which power? Try 'powers'.");return}
	var chosen *power.Power
	powerName:=strings.Join(args," ")
	for _,p:=range power.Unlocked(s.Character.Domain,s.Character.Level) {
		if strings.EqualFold(p.Name,powerName) || strings.EqualFold(strings.ReplaceAll(p.Name," ",""),strings.ReplaceAll(powerName," ","")) {q:=p;chosen=&q;break}
	}
	if chosen==nil {s.WriteLine("That power is not yet awakened.");return}
	if s.Character.Mana<chosen.ManaCost {s.WriteLine("Your divine reserves are exhausted.");return}
	s.Character.Mana-=chosen.ManaCost
	result:=combat.Cast(s.Character.Name,chosen.Name,chosen.Damage,string(chosen.Domain),s.Enemy)
	s.WriteLine("\x1b[1;35m%s\x1b[0m",result.Text)
	if result.Killed {
		xp:=s.Enemy.XP
		s.Character.AddExperience(xp)
		s.WriteLine("\x1b[1;32mDivine victory! +%d XP.\x1b[0m",xp)
		s.Enemy=nil
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
		s.WriteLine("\x1b[1;31mYour mortal life ends here. The Styx waits. But something is wrong...\x1b[0m")
		s.Character.Restore()
		s.Character.RoomID="olympus_gates"
		s.Enemy=nil
		s.WriteLine("You awaken at the Gates of Olympus, restored but shaken.")
		s.WriteLine("The black thread around your sword is still there. It has not forgotten you.")
	}
}

func (s *Session) flee() {
	if s.Enemy==nil || s.Enemy.HP<=0 {s.WriteLine("You are not in combat.");return}
	s.Enemy=nil
	s.WriteLine("You break away from the battle and retreat. Sometimes survival is the wiser path.")
}

func (s *Session) currentWeapon() item.Item {
	if len(s.Character.Inventory)==0{return item.Item{Name:"fists",Tier:item.Common,Damage:3,Kind:"fist"}}
	for _,w:=range s.Character.Inventory { if strings.EqualFold(w.Name,s.Character.Weapon){return w} }
	return s.Character.Inventory[0]
}

func (s *Session) inventory() {
	if len(s.Character.Inventory)==0{s.WriteLine("Your inventory is empty.");return}
	s.WriteLine("\x1b[1;33mInventory\x1b[0m")
	for _,i:=range s.Character.Inventory { s.WriteLine("%s%s %s(+%d)%s",s.color(i.TierColor()),i.TierName(),i.Name,i.Damage,s.color("0")) }
}

func (s *Session) powers() {
	if s.Character.Domain=="" {s.WriteLine("Divinity: dormant. Awaken a domain at level 3.");return}
	s.WriteLine("\x1b[1;35m%s Domain — Divinity %d\x1b[0m",s.Character.Domain,s.Character.Divinity)
	for _,p:=range power.ForDomain(s.Character.Domain) {
		status:="LOCKED"; if s.Character.Level>=p.UnlockLevel {status="UNLOCKED"}
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
	s.WriteLine("Soul legacy: %d memories | %d scars | %d oaths | %d favors | %d curses | %d echoes",len(c.Memories),len(c.Scars),len(c.Oaths),len(c.Favors),len(c.Curses),len(c.Echoes))
}

func (s *Session) rebirth(args []string) {
	if !s.Character.CanRebirth() {s.WriteLine("The Fates have not yet opened the way. Reach level 10.");return}

	if s.Character.Life == 1 {
		if len(args)>0 && args[0]!="modern" {
			s.WriteLine("Your first rebirth leads to the modern world. Later lives will open other horizons.")
			return
		}
		s.Character.Rebirth()
		s.Enemy=nil
		s.WriteLine("\x1b[1;35mThe world falls silent. You cross the Styx.\x1b[0m")
		s.WriteLine("Life 1 ends. Life 2 begins.")
		s.WriteLine("You awaken in the modern era. Athens has become a city of glass, engines, and hidden gods.")
		return
	}

	if len(args)==0 {
		s.WriteLine("\x1b[1;33mREBIRTH PATHS\x1b[0m")
		s.WriteLine("  rebirth ancient — return to Ancient Greece")
		s.WriteLine("  rebirth modern  — return to modern Athens")
		s.WriteLine("  rebirth future  — awaken in Athens after the Last Dawn")
		s.WriteLine("  rebirth lunar   — awaken at the Lunar Oracle")
		s.WriteLine("  rebirth far     — awaken on the Last Shore")
		s.WriteLine("Your choice shapes the next chapter, not the end of your story.")
		return
	}

	type destination struct { era, room, name string }
	paths:=map[string]destination{
		"ancient":{"ancient","olympus_gates","Ancient Greece"},
		"modern":{"modern","modern_crossroads","Modern Athens"},
		"future":{"future","future_city","Athens, After the Last Dawn"},
		"lunar":{"future","future_moon","The Lunar Oracle"},
		"far":{"far","far_era","The Last Shore"},
	}
	p,ok:=paths[args[0]]
	if !ok { s.WriteLine("Unknown path. Try 'rebirth' to see the available horizons."); return }

	old:=s.Character.Life
	s.Character.RebirthTo(p.era,p.room)
	s.Enemy=nil
	s.WriteLine("\x1b[1;35mThe world falls silent. You cross the Styx again.\x1b[0m")
	s.WriteLine("Life %d ends. Life %d begins.",old,s.Character.Life)
	s.WriteLine("You awaken in %s.",p.name)
	s.WriteLine("Your previous lives remain part of you. The era changes; the soul does not.")
}

func (s *Session) color(code string) string { return "\x1b["+code+"m" }
