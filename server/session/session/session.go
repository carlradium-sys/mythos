package session

import (
	"bufio"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"fatewalker/game/character"
	"fatewalker/game/combat"
	"fatewalker/game/help"
	"fatewalker/game/item"
	"fatewalker/game/power"
	"fatewalker/game/progression"
	"fatewalker/game/quest"
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
 RelicWardSpent bool
 RelicStrikeSpent bool
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
	defer s.persist()
	if !s.selectCharacter(scanner) { return }
	s.look()
	s.updateQuests()
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
	case "equip":
		s.equip(parts[1:])
	case "talk", "say":
		s.talk(parts[1:])
	case "choose":
		s.choose(parts[1:])
	case "shop", "wares":
		s.shop()
	case "buy":
		s.buy(parts[1:])
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
	case "rest":
		s.rest()
	case "rebirth":
		s.rebirth(parts[1:])
	case "journal","quest","story":
		s.journal()
	case "quests":
		s.questList()
	case "soul","legacy":
		s.soul()
	case "invoke","gift":
		s.invokeGift()
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
   if err==nil {s.Account,err=s.Accounts.Login(p[1],p[2]); if err==nil{s.WriteLine("Account created.")}}
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
  case "logout": if !s.login(scanner){return false}; continue
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
   s.Character=s.Account.Characters[n-1].Character;s.Character.EnsureLifeGift();s.WriteLine("Welcome back, %s.",s.Character.Name);return true
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

func (s *Session) grantSoulRelic(relic item.Item) bool {
	if !relic.Relic {
		return false
	}
	for _, owned := range s.Character.Inventory {
		if owned.Relic && strings.EqualFold(owned.Name, relic.Name) {
			return false
		}
	}
	s.Character.Inventory = append(s.Character.Inventory, relic)
	s.WriteLine("\x1b[1;35mSoul relic gained: %s. It will endure beyond this life.\x1b[0m", relic.Name)
	return true
}

func (s *Session) awardExperience(amount int) {
	if amount <= 0 {
		return
	}
	oldLevel := s.Character.Level
	s.Character.AddExperience(amount)
	if s.Character.Level > oldLevel {
		s.WriteLine("\x1b[1;35mYour soul surges — Level %d reached!\x1b[0m", s.Character.Level)
		s.WriteLine("Your strength and divine reserves have been restored.")
	}
}

func (s *Session) completeQuest(q quest.Quest) {
	s.awardExperience(q.RewardXP)
	s.Character.SetStoryFlag("quest_completed_" + q.ID)
	switch q.ID {
	case "oracle_whisper":
		s.Character.Memories = append(s.Character.Memories, "Pythia spoke a prophecy meant for a soul that has already crossed the Styx.")
		s.WriteLine("A memory settles into your soul. The Oracle's words will follow you beyond this life.")
	case "river_of_memory":
		s.Character.SetStoryFlag("styx_memory_recovered")
		s.Character.Memories = append(s.Character.Memories, "At the Styx, you recovered a memory the river could not swallow.")
		s.Character.Echoes = append(s.Character.Echoes, "Black water runs beneath the city of glass, waiting for your return.")
		s.grantSoulRelic(item.Item{Name: "Styxglass Shard", Tier: item.Epic, Kind: "relic", Relic: true})
		s.WriteLine("The river leaves an echo inside you. It may answer in a life yet to come.")
	case "oath_across_the_river":
		s.Character.Oaths = append(s.Character.Oaths, "I will carry Pythia's warning beyond the river.")
		s.grantSoulRelic(item.Item{Name: "Oracle's Thread", Tier: item.Epic, Kind: "relic", Relic: true})
		s.WriteLine("A silver thread knots itself around your wrist: the Oracle's warning now guards your soul.")
	case "unwritten_path":
		s.Character.Echoes = append(s.Character.Echoes, "A future that was never foretold burns at the edge of memory.")
		s.grantSoulRelic(item.Item{Name: "Unwritten Ember", Tier: item.Epic, Kind: "relic", Relic: true})
		s.WriteLine("A coal of impossible fire settles in your palm, warm but never consumed.")
	case "delphi_trials":
		s.Character.SetStoryFlag("delphi_trials_complete")
		s.Character.Favors = append(s.Character.Favors, "Pythia's Favor")
		s.grantSoulRelic(item.Item{Name: "Laurel of the Seer", Tier: item.Rare, Kind: "relic", Relic: true})
		s.WriteLine("A living laurel wreath takes root in your memory. The next life will not begin blind.")
	}
	faction, change := questReputationReward(q.ID)
	if faction != "" && change != 0 {
		if s.Character.Reputation == nil {
			s.Character.Reputation = map[string]int{}
		}
		s.Character.Reputation[faction] += change
		direction := "increased"
		if change < 0 {
			direction = "decreased"
		}
		s.WriteLine("Your standing with %s has %s by %d.", faction, direction, abs(change))
	}
}

func questReputationReward(id string) (string, int) {
	switch id {
	case "black_thread":
		return "olympians", 1
	case "river_of_memory":
		return "underworld", 1
	case "oath_across_the_river":
		return "delphi", 1
	case "unwritten_path":
		return "underworld", 1
	case "delphi_trials":
		return "delphi", 1
	default:
		return "", 0
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func (s *Session) ensureQuests() {
 if s.Character.Quests==nil {s.Character.Quests=map[string]int{}}
}
func (s *Session) questUnlocked(q quest.Quest) bool {
	if q.RequiredFlag != "" && !s.Character.HasStoryFlag(q.RequiredFlag) {
		return false
	}
	if q.RequiredQuest != "" && s.Character.Quests[q.RequiredQuest] < 1 {
		return false
	}
	return true
}

func (s *Session) updateQuests() {
 s.ensureQuests()
 for _,q:=range quest.All() {
  if !s.questUnlocked(q) {continue}
  if s.Character.Quests[q.ID]>=q.Required {continue}
  if q.TargetRoom!="" && q.TargetEnemy=="" && s.Character.RoomID==q.TargetRoom {
   old:=s.Character.Quests[q.ID];s.Character.Quests[q.ID]=q.Required
   if old<q.Required {s.WriteLine("\x1b[1;33mQuest advanced: %s\x1b[0m",q.Name)}
   if q.RewardXP>0 {s.completeQuest(q);s.WriteLine("\x1b[1;32mQuest complete! +%d XP.\x1b[0m",q.RewardXP)}
  }
 }
}
func (s *Session) advanceQuestKill(enemy string) {
 s.ensureQuests()
 for _,q:=range quest.All() {
  if !s.questUnlocked(q) {continue}
  if q.TargetEnemy!="" && strings.EqualFold(q.TargetEnemy,enemy) && (q.TargetRoom=="" || q.TargetRoom==s.Character.RoomID) && s.Character.Quests[q.ID]<q.Required {
   s.Character.Quests[q.ID]++
   if s.Character.Quests[q.ID]>=q.Required {s.WriteLine("\x1b[1;33mQuest complete: %s\x1b[0m",q.Name);s.completeQuest(q);s.WriteLine("\x1b[1;32m+%d XP.\x1b[0m",q.RewardXP)}
  }
 }
}
func (s *Session) questList() {
 s.ensureQuests()
 s.WriteLine("\x1b[1;33mQUESTS\x1b[0m")
 for _,q:=range quest.All() {
  p:=s.Character.Quests[q.ID];if p>q.Required{p=q.Required}
  status:="active"
  if p>=q.Required {status="complete"} else if !s.questUnlocked(q) {status="locked"}
  s.WriteLine("  [%s] %s — %d/%d",status,q.Name,p,q.Required)
  s.WriteLine("      %s",q.Goal)
  if q.RewardXP > 0 { s.WriteLine("      Reward: %d XP", q.RewardXP) }
  if status=="locked" {
   if q.RequiredFlag!="" && !s.Character.HasStoryFlag(q.RequiredFlag) {
    s.WriteLine("      Unlock condition: discover %s",strings.ReplaceAll(q.RequiredFlag,"_"," "))
   } else if q.RequiredQuest!="" {
    if prerequisite,ok:=quest.Get(q.RequiredQuest);ok {s.WriteLine("      Unlock condition: complete %s",prerequisite.Name)}
   }
  }
 }
}

func (s *Session) nextAvailableQuest() (quest.Quest, bool) {
	s.ensureQuests()
	for _, q := range quest.All() {
		if s.Character.Quests[q.ID] >= q.Required || !s.questUnlocked(q) {
			continue
		}
		return q, true
	}
	return quest.Quest{}, false
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
	if q, ok := s.nextAvailableQuest(); ok {
		s.WriteLine("\x1b[1;33mSuggested thread: %s\x1b[0m", q.Name)
		s.WriteLine("%s", q.Goal)
		s.WriteLine("Use 'quests' to review every active and locked thread.")
	} else {
		s.WriteLine("No new story thread is calling clearly. Explore, revisit old places, or speak with the people you have met.")
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
	if npcs:=s.World.NPCs[s.Character.RoomID];len(npcs)>0{for _,n:=range npcs{s.WriteLine("\x1b[1;36m%s\x1b[0m — %s",n.Name,n.Description)};s.WriteLine("You can talk <topic> or shop.")}
	if s.Enemy==nil { s.spawnEnemy() }
}

func (s *Session) spawnEnemy() {
	s.RelicWardSpent = false
	s.RelicStrikeSpent = false
	switch s.Character.RoomID {
	case "olympus_foothills":
		s.Enemy = combat.NewHarpy()
	case "oracle_path":
		s.Enemy = combat.NewSatyr()
	case "manticore_den":
		s.Enemy = combat.NewManticore()
	case "modern_plaka":
		s.Enemy = &combat.Enemy{Name: "Street Shade", Description: "A human-shaped shadow slips between the neon signs, moving against the light.", Level: 2, HP: 80, MaxHP: 80, Defense: 5, Damage: 10, XP: 90}
	case "modern_metro":
		s.Enemy = &combat.Enemy{Name: "Echo Hound", Description: "A hound made of static and old memories emerges from the empty train.", Level: 4, HP: 145, MaxHP: 145, Defense: 7, Damage: 15, XP: 170}
	case "modern_styx":
		s.Enemy = &combat.Enemy{Name: "Styx Wraith", Description: "Black water rises into the shape of a veiled figure, carrying coins from lives you never lived.", Level: 5, HP: 185, MaxHP: 185, Defense: 8, Damage: 18, XP: 230}
	case "modern_sanctum":
		s.Enemy = &combat.Enemy{Name: "Bronze Sentinel", Description: "A bronze guardian unfolds from the hidden temple wall, its eyes burning with borrowed starlight.", Level: 6, HP: 220, MaxHP: 220, Defense: 9, Damage: 20, XP: 280}
	case "future_city":
		s.Enemy = &combat.Enemy{Name: "Chronal Warden", Description: "A guardian of fractured seconds steps from a ripple in the air.", Level: 7, HP: 245, MaxHP: 245, Defense: 10, Damage: 22, XP: 320}
	case "future_skyway":
		s.Enemy = &combat.Enemy{Name: "Storm Automaton", Description: "A machine of celestial bronze and lightning blocks the road above the clouds.", Level: 8, HP: 275, MaxHP: 275, Defense: 11, Damage: 25, XP: 370}
	case "future_moon":
		s.Enemy = &combat.Enemy{Name: "Moonshade", Description: "A pale shadow detaches itself from the lunar sanctuary's wall and reaches for your memories.", Level: 9, HP: 305, MaxHP: 305, Defense: 12, Damage: 27, XP: 420}
	case "far_era":
		s.Enemy = &combat.Enemy{Name: "Last Shore Titan", Description: "The drowned world's final guardian rises from the surf, carrying the weight of vanished centuries.", Level: 10, HP: 360, MaxHP: 360, Defense: 13, Damage: 30, XP: 500}
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
	if (strings.HasPrefix(id, "future_") || id == "far_era") && c.Life < 3 {
		s.WriteLine("That horizon belongs to a later chapter of your soul. Rebirth must open the way.")
		return false
	}
	if id=="olympus_road" && c.Level<6 { s.WriteLine("The path into Olympus is veiled by divine law. Become stronger before attempting the ascent."); return false }
	if id=="styx_shore" && c.Level<7 { s.WriteLine("The black river calls, but you are not yet strong enough to cross its threshold."); return false }
	if id=="cerberus_gate" && c.Level<9 { s.WriteLine("A presence beyond the gate warns you away. The guardian is not yet your battle."); return false }
	return true
}

func (s *Session) recordDiscovery(roomID string) {
	flag := "room_discovered_" + roomID
	if s.Character.HasStoryFlag(flag) {
		return
	}
	room := s.World.GetRoom(roomID)
	if room == nil {
		return
	}
	reward := 60
	switch roomID {
	case "delphi_sanctum", "temple_dawn", "styx_shore", "underworld_crossroads", "cerberus_gate", "tartarus_edge", "modern_styx", "modern_sanctum", "future_moon", "far_era":
		reward = 110
	}
	s.Character.SetStoryFlag(flag)
	s.WriteLine("Discovery: %s. +%d exploration XP.", room.Name, reward)
	s.awardExperience(reward)
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
	s.recordDiscovery(next)
	s.look()
	s.updateQuests()
	if s.Character.RoomID=="olympus_foothills" { s.advanceTutorial(2) }
}

func (s *Session) attack() {
	if s.Enemy == nil || s.Enemy.HP <= 0 {
		s.WriteLine("There is nothing here to fight.")
		return
	}
	w := s.currentWeapon()
	result := combat.Attack(s.Character.Name, w, s.Enemy, s.Character.AttackPower())
	if result.Damage > 0 {
		totalDamage, bonus := s.applySoulRelicStrike(result.Damage)
		if bonus > 0 {
			s.Enemy.HP -= bonus
			if s.Enemy.HP < 0 {
				s.Enemy.HP = 0
			}
			result.Damage = totalDamage
			result.Killed = s.Enemy.HP == 0
			result.Text += fmt.Sprintf("\nThe %s answers your strike, adding %d damage. (%d total damage)", s.soulRelicStrikeName(), bonus, totalDamage)
			if result.Killed {
				result.Text += " The " + s.Enemy.Name + " collapses."
			}
		}
	}
	s.WriteLine("%s", result.Text)
	s.advanceTutorial(3)
	if result.Killed {
		s.defeatEnemy(false)
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
		s.defeatEnemy(true)
		return
	}
	s.enemyTurn()
}

func (s *Session) defeatEnemy(divine bool) {
	if s.Enemy == nil {
		return
	}
	enemy := s.Enemy
	s.advanceQuestKill(enemy.Name)
	s.awardExperience(enemy.XP)
	if divine {
		s.WriteLine("\x1b[1;32mDivine victory! +%d XP.\x1b[0m", enemy.XP)
	} else {
		s.WriteLine("\x1b[1;32mVictory! +%d XP.\x1b[0m", enemy.XP)
	}
	goldReward := 5 + enemy.Level*5
	s.Character.Gold += goldReward
	s.WriteLine("You recover %d drachmae. Gold: %d.", goldReward, s.Character.Gold)
	if enemy.Name == "Manticore" && s.Character.Level >= 5 {
		hasFang := false
		for _, owned := range s.Character.Inventory {
			if strings.EqualFold(owned.Name, "manticore fang") {
				hasFang = true
				break
			}
		}
		if !hasFang {
			drop := item.Item{Name: "manticore fang", Tier: item.Rare, Damage: 18, Kind: "sword"}
			s.Character.Inventory = append(s.Character.Inventory, drop)
			s.WriteLine("%sYou recover a %s %s%s.", s.color(drop.TierColor()), drop.TierName(), drop.Name, s.color("0"))
			s.WriteLine("The weapon hums faintly, as if it remembers the creature.")
		}
	}
	s.Enemy = nil
	s.RelicWardSpent = false
	s.RelicStrikeSpent = false
}

func (s *Session) applySoulRelicWard(damage int) (int, int) {
	if damage <= 0 || s.RelicWardSpent {
		return damage, 0
	}
	wardStrength := 0
	for _, owned := range s.Character.Inventory {
		if !owned.Relic {
			continue
		}
		switch strings.ToLower(owned.Name) {
		case "styxglass shard":
			if wardStrength < 8 { wardStrength = 8 }
		case "oracle's thread":
			if wardStrength < 4 { wardStrength = 4 }
		}
	}
	if wardStrength == 0 {
		return damage, 0
	}
	absorbed := wardStrength
	if absorbed > damage { absorbed = damage }
	s.RelicWardSpent = true
	return damage - absorbed, absorbed
}

func (s *Session) soulRelicStrikeName() string {
	// Prefer the strongest offensive relic so inventory order never changes combat results.
	for _, name := range []string{"Unwritten Ember", "Laurel of the Seer"} {
		for _, owned := range s.Character.Inventory {
			if owned.Relic && strings.EqualFold(owned.Name, name) {
				return name
			}
		}
	}
	return "soul relic"
}

func (s *Session) applySoulRelicStrike(damage int) (int, int) {
	if damage <= 0 || s.RelicStrikeSpent {
		return damage, 0
	}
	bonus := 0
	if s.soulRelicStrikeName() == "Unwritten Ember" {
		bonus = 6
	} else if s.soulRelicStrikeName() == "Laurel of the Seer" {
		bonus = 3
	}
	if bonus == 0 {
		return damage, 0
	}
	s.RelicStrikeSpent = true
	return damage + bonus, bonus
}

func (s *Session) enemyTurn() {
	if s.Enemy==nil || s.Enemy.HP<=0{return}
	result:=combat.EnemyAttack(s.Enemy,s.Character.DefensePower())
	if result.Damage>0 {
		finalDamage, absorbed := s.applySoulRelicWard(result.Damage)
		if absorbed > 0 {
			result.Damage = finalDamage
			result.Text += fmt.Sprintf("\nA soul relic flashes; its ward absorbs %d damage.", absorbed)
		}
		s.Character.HP-=result.Damage
		if s.Character.HP<0{s.Character.HP=0}
	}
	s.WriteLine("%s", result.Text)
	s.WriteLine("\x1b[1;31m%s HP: %d/%d\x1b[0m  |  \x1b[1;36m%s HP: %d/%d\x1b[0m", s.Enemy.Name, s.Enemy.HP, s.Enemy.MaxHP, s.Character.Name, s.Character.HP, s.Character.MaxHP)
	if s.Character.HP == 0 {
		s.WriteLine("\x1b[1;31mYour strength fails. Black water closes over your vision, and the Styx refuses to keep you.\x1b[0m")
		s.Character.Restore()
		s.Character.RoomID = "olympus_gates"
		s.Enemy = nil
		s.RelicWardSpent = false
		s.RelicStrikeSpent = false
		s.WriteLine("You awaken at the Gates of Olympus, restored but shaken. This life continues.")
		s.WriteLine("The black thread around your sword is still there. It has not forgotten you.")
	}
}

func (s *Session) rest() {
	if s.Enemy != nil && s.Enemy.HP > 0 {
		s.WriteLine("You cannot rest while %s is still hunting you.", s.Enemy.Name)
		return
	}
	oldHP, oldMana := s.Character.HP, s.Character.Mana
	s.Character.Restore()
	s.WriteLine("You take a moment to gather yourself. Restored %d health and %d mana.", s.Character.HP-oldHP, s.Character.Mana-oldMana)
}

func (s *Session) flee() {
	if s.Enemy==nil || s.Enemy.HP<=0 {s.WriteLine("You are not in combat.");return}
	s.Enemy=nil
	s.RelicWardSpent = false
	s.RelicStrikeSpent = false
	s.WriteLine("You break away from the battle and retreat. Sometimes survival is the wiser path.")
}

func (s *Session) currentWeapon() item.Item {
	if len(s.Character.Inventory)==0{return item.Item{Name:"fists",Tier:item.Common,Damage:3,Kind:"fist"}}
	for _,w:=range s.Character.Inventory { if strings.EqualFold(w.Name,s.Character.Weapon){return w} }
	return s.Character.Inventory[0]
}

func (s *Session) equip(args []string) {
 if len(args)==0 {s.WriteLine("Equip what?");return}
 name:=strings.Join(args," ");for _,i:=range s.Character.Inventory {
  if strings.EqualFold(i.Name,name) {
   if item.IsWeapon(i) {s.Character.Weapon=i.Name;s.WriteLine("You equip %s.",i.Name);return}
   if i.Kind=="armor" {s.Character.Armor=i.Name;s.WriteLine("You wear %s. Defense +%d.",i.Name,i.Armor);return}
  }
 }
 s.WriteLine("You do not possess that equipment.")
}

func (s *Session) inventory() {
	if len(s.Character.Inventory)==0{s.WriteLine("Your inventory is empty.");return}
	s.WriteLine("\x1b[1;33mInventory\x1b[0m")
	for _,i:=range s.Character.Inventory { equipped:=""; if strings.EqualFold(i.Name,s.Character.Weapon)||strings.EqualFold(i.Name,s.Character.Armor){equipped=" [equipped]"}; relic:=""; if i.Relic { relic=" [soul relic]" }; s.WriteLine("%s%s %s(+%d armor/%d damage)%s%s%s",s.color(i.TierColor()),i.TierName(),i.Name,i.Armor,i.Damage,equipped,relic,s.color("0")) }
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
	s.WriteLine("Attack: %d   Defense: %d   Armor: %d   Divinity: %d   Domain: %s   Gold: %d",c.AttackPower(),c.DefensePower(),c.ArmorPower(),c.Divinity,c.Domain,c.Gold)
	s.WriteLine("Soul legacy: %d memories | %d scars | %d oaths | %d favors | %d curses | %d echoes",len(c.Memories),len(c.Scars),len(c.Oaths),len(c.Favors),len(c.Curses),len(c.Echoes))
}

func soulRelicCount(inventory []item.Item) int {
	count := 0
	for _, owned := range inventory {
		if owned.Relic {
			count++
		}
	}
	return count
}

func soulRelicEffect(name string) string {
	switch strings.ToLower(name) {
	case "styxglass shard":
		return "once per encounter, absorbs up to 8 incoming damage"
	case "oracle's thread":
		return "once per encounter, absorbs up to 4 incoming damage"
	case "unwritten ember":
		return "adds 6 damage to the first successful weapon strike each encounter"
	case "laurel of the seer":
		return "adds 3 damage to the first successful weapon strike each encounter"
	default:
		return "its deeper purpose is still unknown"
	}
}

func (s *Session) soul() {
	c := s.Character
	s.WriteLine("\x1b[1;35mSOUL LEGACY\x1b[0m")
	s.WriteLine("Life %d — %s. These marks belong to the soul, not just this era.", c.Life, c.Era)
	giftStatus := "ready"
	if c.LifeGiftUsed {
		giftStatus = "spent until rebirth"
	}
	s.WriteLine("Life-gift: %s (%s). Use 'invoke' to call on it.", c.LifeGift, giftStatus)
	s.WriteLine("\x1b[1;33mSoul relics (%d)\x1b[0m", soulRelicCount(c.Inventory))
	hasRelic := false
	for _, owned := range c.Inventory {
		if owned.Relic {
			hasRelic = true
			s.WriteLine("  %s — %s", owned.Name, soulRelicEffect(owned.Name))
		}
	}
	if !hasRelic {
		s.WriteLine("  No soul relics recovered yet.")
	}
	s.WriteLine("\x1b[1;33mFaction standing\x1b[0m")
	if c.Reputation == nil {
		c.Reputation = map[string]int{}
	}
	factions := []string{"delphi", "olympians", "underworld"}
	for _, faction := range factions {
		s.WriteLine("  %-12s %d", faction, c.Reputation[faction])
	}
	printLegacy := func(label string, entries []string) {
		s.WriteLine("\x1b[1;33m%s (%d)\x1b[0m", label, len(entries))
		if len(entries) == 0 {
			s.WriteLine("  Nothing recorded yet.")
			return
		}
		for _, entry := range entries {
			s.WriteLine("  • %s", entry)
		}
	}
	printLegacy("Memories", c.Memories)
	printLegacy("Oaths", c.Oaths)
	printLegacy("Scars", c.Scars)
	printLegacy("Favors", c.Favors)
	printLegacy("Curses", c.Curses)
	printLegacy("Echoes", c.Echoes)
	s.WriteLine("\x1b[1;33mStory choices\x1b[0m")
	flags := make([]string, 0, len(c.StoryFlags))
	for flag, chosen := range c.StoryFlags {
		if chosen && !strings.HasPrefix(flag, "room_discovered_") {
			flags = append(flags, flag)
		}
	}
	if len(flags) == 0 {
		s.WriteLine("  No pivotal choices recorded yet.")
		return
	}
	sort.Strings(flags)
	for _, flag := range flags {
		s.WriteLine("  %s", strings.ReplaceAll(flag, "_", " "))
	}
}

func (s *Session) invokeGift() {
	c := s.Character
	c.EnsureLifeGift()
	if c.LifeGift == "" {
		s.WriteLine("Your soul has no manifested life-gift yet.")
		return
	}
	if c.LifeGiftUsed {
		s.WriteLine("%s has already answered you in this life. The gift will return after rebirth.", c.LifeGift)
		return
	}

	switch c.LifeGift {
	case "Thread Sense":
		room := s.World.GetRoom(c.RoomID)
		if room == nil {
			s.WriteLine("The thread finds no stable place to reveal.")
			return
		}
		exits := make([]string, 0, len(room.Exits))
		for direction := range room.Exits {
			exits = append(exits, direction)
		}
		sort.Strings(exits)
		s.WriteLine("The black thread tightens. You sense %s.", room.Name)
		if len(exits) > 0 {
			s.WriteLine("Possible paths: %s.", strings.Join(exits, ", "))
		}
		active := 0
		for _, q := range quest.All() {
			if s.questUnlocked(q) && c.Quests[q.ID] < q.Required {
				active++
			}
		}
		s.WriteLine("You sense %d available story thread(s). Your choice of path remains yours.", active)
	case "Echo Sight":
		recovered := 0
		if c.MaxMana > c.Mana {
			recovered = 12
			if c.Mana+recovered > c.MaxMana {
				recovered = c.MaxMana - c.Mana
			}
			c.Mana += recovered
		}
		s.WriteLine("The modern world briefly overlays every life you have lived. You recover %d mana.", recovered)
		if len(c.Memories) > 0 {
			s.WriteLine("Memory: %s", c.Memories[len(c.Memories)-1])
		}
		if len(c.Echoes) > 0 {
			s.WriteLine("Echo: %s", c.Echoes[len(c.Echoes)-1])
		}
	case "Chronal Pulse":
		before := c.HP
		c.HP += 25
		if c.HP > c.MaxHP {
			c.HP = c.MaxHP
		}
		s.WriteLine("Time folds around your wounds. You recover %d health.", c.HP-before)
	case "Moon's Shelter":
		oldHP, oldMana := c.HP, c.Mana
		c.HP += 20
		if c.HP > c.MaxHP { c.HP = c.MaxHP }
		c.Mana += 10
		if c.Mana > c.MaxMana { c.Mana = c.MaxMana }
		s.WriteLine("Silver light gathers around you. You recover %d health and %d mana.", c.HP-oldHP, c.Mana-oldMana)
	case "Fateweaver's Knot":
		if len(c.Curses) > 0 {
			curse := c.Curses[len(c.Curses)-1]
			c.Curses = c.Curses[:len(c.Curses)-1]
			s.WriteLine("You pull one strand loose from the Fates' knot. The curse fades: %s.", curse)
		} else {
			favor := "The Fates granted you a second chance at the Last Shore."
			c.Favors = append(c.Favors, favor)
			s.WriteLine("The knot loosens. With no curse to sever, the Fates leave you a favor.")
		}
	default:
		s.WriteLine("Your life-gift has not yet learned how to answer.")
		return
	}
	c.LifeGiftUsed = true
	s.WriteLine("Life-gift used: %s. It will return after your next rebirth.", c.LifeGift)
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

func(s *Session) currentNPC()*world.NPC{n:=s.World.NPCs[s.Character.RoomID];if len(n)==0{return nil};return n[0]}
func (s *Session) factionGreeting(n *world.NPC) string {
	if n == nil || n.Faction == "" || s.Character.Reputation == nil {
		return ""
	}
	switch standing := s.Character.Reputation[n.Faction]; {
	case standing >= 3:
		return fmt.Sprintf("%s greets you with newfound warmth. Your deeds have earned the faction's respect.", n.Name)
	case standing <= -2:
		return fmt.Sprintf("%s regards you warily. Your standing with the %s has made trust difficult.", n.Name, n.Faction)
	default:
		return ""
	}
}

func(s *Session) talk(args []string) {
	n := s.currentNPC()
	if n == nil {
		s.WriteLine("There is no one here willing to speak with you.")
		return
	}
	topic := "hello"
	if len(args) > 0 {
		topic = strings.Join(args, " ")
	}
	if greeting := s.persistentNPCGreeting(n, topic); greeting != "" {
		s.WriteLine("%s", greeting)
		return
	}
	if topic == "hello" || topic == "greeting" {
		if greeting := s.factionGreeting(n); greeting != "" {
			s.WriteLine("%s", greeting)
			return
		}
	}
	if n.ID == "pythia" && (topic == "hello" || topic == "greeting") {
		switch {
		case s.Character.HasStoryFlag("oracle_trust"):
			s.WriteLine("\x1b[1;36mPythia smiles faintly. \"You kept the thread I gave you. The river will test that promise when you least expect it.\"\x1b[0m")
			return
		case s.Character.HasStoryFlag("oracle_defied"):
			s.WriteLine("\x1b[1;36mPythia regards you without anger. \"Still walking your own road, I see. Even the Fates have learned to leave a little room for defiance.\"\x1b[0m")
			return
		}
	}
	if topic == "choice" || topic == "choices" {
		if n.ID == "pythia" {
			if s.Character.HasStoryFlag("oracle_choice_made") {
				s.WriteLine("Pythia studies you. \"The river has recorded your answer. You cannot make that choice unmade.\"")
				return
			}
			s.WriteLine("Pythia's voice falls to a whisper: \"When the Fates offer a thread, will you trust the pattern or cut your own path?\"")
			s.WriteLine("  choose trust — accept the Oracle's guidance and swear to remember it.")
			s.WriteLine("  choose defy  — reject prophecy and bear the consequences alone.")
			return
		}
	}
	if t := n.DialogueFor(topic); t != "" {
		s.WriteLine("\x1b[1;36m%s:\x1b[0m %s", n.Name, t)
		return
	}
	s.WriteLine("\x1b[1;36m%s:\x1b[0m \"Ask me about the things that matter here.\"", n.Name)
}


func (s *Session) persistentNPCGreeting(n *world.NPC, topic string) string {
	if n == nil || n.ID != "athens_vendor" || (topic != "hello" && topic != "greeting") {
		return ""
	}
	if s.Character.HasStoryFlag("styx_memory_recovered") {
		return "Myrto's smile fades as she studies you. \"You found the river beneath the old world. I wondered when it would recognize you here.\""
	}
	if s.Character.HasStoryFlag("oracle_trust") {
		return "Myrto studies the oath-mark at your wrist. \"Someone from Delphi trusted you with a warning. In this city, promises have a longer shadow than gods expect.\""
	}
	if s.Character.HasStoryFlag("oracle_defied") {
		return "Myrto notices the faint scar in your aura. \"You told an Oracle no and survived it? Good. Athens has enough people who mistake prophecy for permission.\""
	}
	if s.Character.HasStoryFlag("quest_completed_oracle_whisper") {
		return "Myrto tilts her head. \"You carry an old prophecy. In this city, old words have a way of becoming new trouble.\""
	}
	return ""
}

func (s *Session) choose(args []string) {
	if len(args) == 0 {
		s.WriteLine("Choose what? At the Oracle, try 'talk choices' first.")
		return
	}
	n := s.currentNPC()
	if n == nil || n.ID != "pythia" || s.Character.RoomID != "oracle_path" {
		s.WriteLine("There is no choice here for you to make.")
		return
	}
	if s.Character.HasStoryFlag("oracle_choice_made") {
		s.WriteLine("Pythia shakes her head. \"The river has recorded your answer. The choice belongs to your soul now.\"")
		return
	}
	if s.Character.StoryFlags == nil {
		s.Character.StoryFlags = map[string]bool{}
	}
	if s.Character.Reputation == nil {
		s.Character.Reputation = map[string]int{}
	}
	switch strings.ToLower(strings.Join(args, " ")) {
	case "trust", "trust oracle", "accept":
		s.Character.SetStoryFlag("oracle_choice_made")
		s.Character.SetStoryFlag("oracle_trust")
		s.Character.Reputation["delphi"] += 2
		s.Character.Oaths = append(s.Character.Oaths, "I will remember the Oracle's warning when the Styx asks what I would surrender.")
		s.Character.Memories = append(s.Character.Memories, "Pythia offered a path through uncertainty, and I chose to listen.")
		s.WriteLine("\x1b[1;36mPythia lowers her head. \"Then carry my words beyond this life. Trust is not obedience; it is a promise to remember.\"\x1b[0m")
		s.WriteLine("Delphi reputation +2. A new oath has taken root in your soul.")
	case "defy", "defy oracle", "reject":
		s.Character.SetStoryFlag("oracle_choice_made")
		s.Character.SetStoryFlag("oracle_defied")
		s.Character.Reputation["delphi"] -= 1
		s.Character.Scars = append(s.Character.Scars, "I refused the Oracle's offered path; the future must answer to my own hand.")
		s.Character.Echoes = append(s.Character.Echoes, "A prophecy ended at the moment I refused to hear its ending.")
		s.WriteLine("\x1b[1;36mPythia's brazier gutters. \"Then walk without my blessing. Even defiance leaves a thread behind.\"\x1b[0m")
		s.WriteLine("Delphi reputation -1. Your refusal leaves a scar that will follow you.")
	default:
		s.WriteLine("Pythia waits. Choose 'trust' or 'defy'.")
	}
}
func (s *Session) merchantPrice(n *world.NPC, item item.Item) int {
	price := item.Price
	if n == nil || n.Faction == "" || s.Character.Reputation == nil {
		return price
	}
	switch standing := s.Character.Reputation[n.Faction]; {
	case standing >= 3:
		price = price * 90 / 100
	case standing <= -2:
		price = (price*110 + 99) / 100
	}
	if price < 1 && item.Price > 0 {
		price = 1
	}
	return price
}

func(s *Session) shop(){n:=s.currentNPC();if n==nil||!n.HasShop(){s.WriteLine("There is no merchant here.");return};s.WriteLine("\x1b[1;33m%s's WARES\x1b[0m — You have %d gold",n.Name,s.Character.Gold);for _,i:=range n.Shop{price:=s.merchantPrice(n,i);note:="";if price<i.Price{note=" (faction discount)"}else if price>i.Price{note=" (faction surcharge)"};s.WriteLine("  %s%s %s — %d gold%s",s.color(i.TierColor()),i.TierName(),i.Name,price,note)};s.WriteLine("Use: buy <item>")}
func(s *Session) buy(args []string){if len(args)==0{s.WriteLine("Buy what? Try 'shop'.");return};n:=s.currentNPC();if n==nil||!n.HasShop(){s.WriteLine("There is no merchant here.");return};name:=strings.Join(args," ");for _,o:=range n.Shop{if strings.EqualFold(o.Name,name){price:=s.merchantPrice(n,o);if s.Character.Gold<price{s.WriteLine("You need %d more gold.",price-s.Character.Gold);return};for _,owned:=range s.Character.Inventory{if strings.EqualFold(owned.Name,o.Name){s.WriteLine("You already possess that item.");return}};s.Character.Gold-=price;s.Character.Inventory=append(s.Character.Inventory,o);s.WriteLine("\x1b[1;32mPurchased %s for %d gold.\x1b[0m",o.Name,price);return}};s.WriteLine("That item is not for sale here.")}
