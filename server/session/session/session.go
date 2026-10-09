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
 EncounterClearedRoom string
 PendingEquip []string
	AutoMap bool
}

func New(accounts *account.Store, conn net.Conn, w *world.World) *Session {
 return &Session{Accounts:accounts, Conn:conn, World:w, AutoMap:false}
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
	s.prepareTutorial()
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
	case "tutorial", "guide":
		s.tutorial()
	case "look","l":
		s.look()
		if s.TutorialStep == 1 {
			s.advanceTutorial(2)
		}
	case "map":
		s.localMap()
	case "automap", "mapauto", "maptoggle":
		s.autoMap(parts[1:])
	case "worldmap", "world-map":
		s.WriteLine(s.World.MapText(s.Character.RoomID))
	case "exits":
		s.showExits()
	case "north","south","east","west","up","down","in","out","n","s","e","w","u","d":
		s.move(parts[0])
	case "who":
		s.WriteLine("You are the first known traveler on this shard.")
	case "score","stats":
		s.score()
	case "inventory", "inv", "i":
		s.inventory()
	case "equipment", "eq", "gear":
		s.equipment()
	case "equip", "wield", "wear":
		s.equip(parts[1:])
	case "talk", "say":
		s.talk(parts[1:])
	case "hint":
		s.hint(parts[1:])
	case "choose":
		s.choose(parts[1:])
	case "shop", "wares":
		s.shop(parts[1:])
	case "buy":
		s.buy(parts[1:])
	case "attack","kill","hit":
		s.attack(parts[1:])
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
	case "museum_echo":
		s.Character.SetStoryFlag("museum_echo_uncovered")
		s.Character.Memories = append(s.Character.Memories, "A museum artifact remembered you before you remembered it.")
		s.Character.Echoes = append(s.Character.Echoes, "Behind glass, an ancient face wears your eyes.")
		s.WriteLine("The artifact warms beneath the glass. A memory crosses the distance between lives.")
	case "gate_of_three":
		s.Character.SetStoryFlag("cerberus_defeated")
		s.Character.Favors = append(s.Character.Favors, "Cerberus Oath")
		s.WriteLine("The guardian falls silent. A new favor settles into your soul.")
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
	case "gate_of_three":
		return "underworld", 2
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
	s.WriteLine("")
	s.WriteLine("\x1b[1;36m╭─ STORY THREADS / QUESTS ─╮\x1b[0m")
	active, complete, locked := 0, 0, 0
	for _, q := range quest.All() {
		p := s.Character.Quests[q.ID]
		if p > q.Required { p = q.Required }
		status, color := "ACTIVE", "\x1b[1;33m"
		if p >= q.Required {
			status, color = "COMPLETE", "\x1b[1;32m"
			complete++
		} else if !s.questUnlocked(q) {
			status, color = "LOCKED", "\x1b[90m"
			locked++
		} else {
			active++
		}
		s.WriteLine("")
		s.WriteLine("  %s[%s]\x1b[0m %s  (%d/%d)", color, status, q.Name, p, q.Required)
		s.WriteLine("    %s", q.Goal)
		if status == "ACTIVE" && q.Required > 1 {
			s.WriteLine("    Progress: %d of %d", p, q.Required)
		}
		if q.RewardXP > 0 && status != "LOCKED" {
			s.WriteLine("    Reward: %d XP", q.RewardXP)
		}
		if status == "LOCKED" {
			if q.RequiredFlag != "" && !s.Character.HasStoryFlag(q.RequiredFlag) {
				s.WriteLine("    Unlock: discover %s", strings.ReplaceAll(q.RequiredFlag, "_", " "))
			} else if q.RequiredQuest != "" {
				if prerequisite, ok := quest.Get(q.RequiredQuest); ok {
					s.WriteLine("    Unlock: complete %s", prerequisite.Name)
				}
			}
		}
	}
	s.WriteLine("")
	s.WriteLine("\x1b[90mSummary: %d active  •  %d complete  •  %d locked\x1b[0m", active, complete, locked)
	s.WriteLine("Type 'journal' for a concise suggestion on what to do next.")
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
	s.WriteLine("You began in Asterion, a quiet village beneath the mountain.")
	s.WriteLine("A black thread binds itself to your bronze sword.")
	s.WriteLine("Something beyond the village knows your name.")
	s.WriteLine("")
	s.WriteLine("\x1b[1;33mCurrent thread\x1b[0m")
	switch {
	case s.Character.Life == 1 && s.Character.Level < 3:
		s.WriteLine("Complete the village tutorial, then explore the foothills and discover why the creatures seem to recognize you.")
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

func (s *Session) tutorial() {
	s.WriteLine("\x1b[1;33mTHE FIRST THREAD — GUIDED TUTORIAL\x1b[0m")
	s.tutorialHint()
	s.WriteLine("Useful later: inv, eq, score, quests, journal, help <topic>, map, and hint.")
	s.WriteLine("The tutorial is saved with your character. You can explore freely after completing these first lessons.")
}

func (s *Session) prepareTutorial() {
	if s.Character == nil {
		return
	}
	if s.Character.Life == 1 && s.Character.Level <= 1 &&
		s.Character.RoomID == "olympus_gates" &&
		!s.Character.HasStoryFlag("room_discovered_olympus_foothills") &&
		!s.Character.HasStoryFlag("tutorial_complete") {
		s.Character.RoomID = "village_square"
	}
	switch {
	case s.Character.HasStoryFlag("tutorial_complete"):
		s.TutorialStep = 4
	case s.Character.HasStoryFlag("tutorial_conversation_complete"):
		s.TutorialStep = 3
	case s.Character.HasStoryFlag("tutorial_look_complete"):
		s.TutorialStep = 2
	case s.Character.HasStoryFlag("tutorial_movement_complete"):
		s.TutorialStep = 1
	default:
		s.TutorialStep = 0
	}
}

func (s *Session) tutorialHint() {
	switch s.TutorialStep {
	case 0:
		s.WriteLine("\x1b[1;36mTUTORIAL 1/4 — MOVEMENT:\x1b[0m Type 'north' to walk from the village square into the village lane.")
	case 1:
		s.WriteLine("\x1b[1;36mTUTORIAL 2/4 — LOOK:\x1b[0m Type 'look' to inspect the lane, its exit, and the village guide.")
	case 2:
		s.WriteLine("\x1b[1;36mTUTORIAL 3/4 — CONVERSATION:\x1b[0m Speak to Damon, the guide. Type 'talk 1 hello'.")
	case 3:
		s.WriteLine("\x1b[1;36mTUTORIAL 4/4 — ASK FOR A HINT:\x1b[0m Type 'hint', then ask Damon about the road with 'talk 1 road'.")
	case 4:
		s.WriteLine("\x1b[1;32mTUTORIAL COMPLETE:\x1b[0m The northern road leads to the foothills. Follow it when ready; use 'hint' whenever an NPC's next step is unclear.")
	}
}

func (s *Session) advanceTutorial(step int) {
	if step <= s.TutorialStep {
		return
	}
	s.TutorialStep = step
	switch step {
	case 1:
		s.Character.SetStoryFlag("tutorial_movement_complete")
	case 2:
		s.Character.SetStoryFlag("tutorial_look_complete")
	case 3:
		s.Character.SetStoryFlag("tutorial_conversation_complete")
	case 4:
		s.Character.SetStoryFlag("tutorial_complete")
	}
	s.tutorialHint()
}

func (s *Session) hint(args []string) {
	index, _ := s.parseNPCSelection(args)
	n := s.currentNPC(index)
	if n == nil {
		s.WriteLine("There is no one here to ask. Use 'look' to find nearby people.")
		return
	}
	if s.Enemy != nil && s.Enemy.HP > 0 {
		s.WriteLine("The %s is still a threat. Defeat it or use 'flee' before speaking.", s.Enemy.Name)
		return
	}
	switch n.ID {
	case "village_guide":
		switch s.TutorialStep {
		case 0:
			s.WriteLine("Damon's hint: learn to move first. Type 'north' to enter the village lane.")
		case 1:
			s.WriteLine("Damon's hint: type 'look' to inspect your surroundings and find the person you can speak with.")
		case 2:
			s.WriteLine("Damon's hint: begin a conversation with 'talk 1 hello'.")
		case 3:
			s.WriteLine("Damon's hint: ask about the road. Type 'talk 1 road'.")
		default:
			s.WriteLine("Damon's hint: the northern road leads to the foothills. The guide can answer 'road' or 'village'.")
		}
	case "pythia":
		if !s.Character.HasStoryFlag("oracle_choice_made") {
			s.WriteLine("Pythia's hint: ask about the choice with 'talk 1 choices', then decide with 'choose trust' or 'choose defy'.")
		} else if !s.Character.HasStoryFlag("quest_completed_oracle_whisper") {
			s.WriteLine("Pythia's hint: ask about 'trials'. The Delphi Sanctum lies farther along the path.")
		} else if s.Character.HasStoryFlag("oracle_trust") && !s.Character.HasStoryFlag("quest_completed_oath_across_the_river") {
			s.WriteLine("Pythia's hint: ask about the 'styx' and follow the river's trail in your journal.")
		} else {
			s.WriteLine("Pythia's hint: review 'quests' and 'journal' for the next unlocked story thread.")
		}
	case "hephaestus_apprentice":
		if s.Character.Quests["black_thread"] < 1 {
			s.WriteLine("Theron's hint: ask about the 'thread', then deal with the Harpy troubling the foothills. Use 'attack 1' when ready.")
		} else {
			s.WriteLine("Theron's hint: your weapon is only the beginning. Check 'inv' and 'shop 1', then follow 'journal' toward Delphi.")
		}
	case "athens_vendor":
		if s.Character.HasStoryFlag("styx_memory_recovered") && !s.Character.HasStoryFlag("museum_echo_uncovered") {
			s.WriteLine("Myrto's hint: ask about the 'museum'. Something old is waiting behind glass.")
		} else {
			s.WriteLine("Myrto's hint: ask about the 'ancient' past, then use 'quests' and 'journal' to follow the memory you recognize.")
		}
	default:
		if q, ok := s.nextAvailableQuest(); ok {
			s.WriteLine("%s seems to be connected to a story thread.", n.Name)
			s.WriteLine("Hint: %s", q.Goal)
			s.WriteLine("Try asking about a topic you noticed in the room description, or type 'quests' for the current thread.")
		} else {
			s.WriteLine("%s has no new hint right now. Try 'talk 1 hello', 'quests', or 'journal'.", n.Name)
		}
	}
}


func (s *Session) autoMap(args []string) {
	if len(args) == 0 {
		state := "ON"
		if !s.AutoMap { state = "OFF" }
		s.WriteLine("")
		s.WriteLine("\x1b[1;36m╭─ AUTOMATIC MAP ─╮\x1b[0m")
		s.WriteLine("  Automatic map after movement: %s", state)
		s.WriteLine("  Usage: automap on | automap off | automap toggle")
		return
	}
	switch strings.ToLower(args[0]) {
	case "on", "yes", "true":
		s.AutoMap = true
	case "off", "no", "false":
		s.AutoMap = false
	case "toggle":
		s.AutoMap = !s.AutoMap
	default:
		s.WriteLine("Use 'automap on', 'automap off', or 'automap toggle'.")
		return
	}
	state := "ON"
	if !s.AutoMap { state = "OFF" }
	s.WriteLine("Automatic map after movement is now %s.", state)
	if s.AutoMap {
		s.localMap()
	}
}

func (s *Session) showExits() {
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil {
		s.WriteLine("You are nowhere. The world has lost track of you.")
		return
	}
	directions := []string{"north", "south", "east", "west", "up", "down", "in", "out"}
	s.WriteLine("")
	s.WriteLine("\x1b[1;36m╭─ PATHS FROM %s ─╮\x1b[0m", r.Name)
	found := false
	for _, direction := range directions {
		if destinationID, ok := r.Exits[direction]; ok {
			found = true
			destinationName := destinationID
			if destination := s.World.GetRoom(destinationID); destination != nil {
				destinationName = destination.Name
			}
			s.WriteLine("  \x1b[36m%-5s →\x1b[0m %s", strings.ToUpper(direction), destinationName)
		}
	}
	if !found {
		s.WriteLine("  No visible paths lead away from here.")
	}
}
func (s *Session) localMap() {
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil {
		s.WriteLine("You are nowhere. The world has lost track of you.")
		return
	}
	destName := func(direction string) string {
		id, ok := r.Exits[direction]
		if !ok { return "" }
		if destination := s.World.GetRoom(id); destination != nil { return destination.Name }
		return id
	}
	north, south := destName("north"), destName("south")
	east, west := destName("east"), destName("west")
	up, down := destName("up"), destName("down")
	in, out := destName("in"), destName("out")

	s.WriteLine("")
	s.WriteLine("\x1b[1;36m╭─ LOCAL MAP ─╮\x1b[0m")
	// Keep the compass grid compact even when room names are long. Full names
	// are listed below so truncation in the diagram never hides information.
	cell := func(name string) string {
		label := name
		runes := []rune(label)
		if len(runes) > 18 {
			label = string(runes[:17]) + "…"
		}
		return fmt.Sprintf("%-20s", "["+label+"]")
	}
	if north != "" {
		s.WriteLine("                          ↑")
		s.WriteLine("                %s", cell(north))
	}
	left, right := "                    ", "                    "
	if west != "" { left = cell(west) }
	if east != "" { right = cell(east) }
	s.WriteLine("%s%s%s", left, "← \x1b[1;32m★ YOU ★\x1b[0m → ", right)
	s.WriteLine("                %s", cell(r.Name))
	if south != "" {
		s.WriteLine("                          ↓")
		s.WriteLine("                %s", cell(south))
	}
	if len(r.Exits) > 0 {
		s.WriteLine("")
		s.WriteLine("\x1b[1;37mCONNECTED ROOMS\x1b[0m")
		for _, direction := range []string{"north", "west", "east", "south"} {
			if name := destName(direction); name != "" {
				s.WriteLine("  %-5s → %s", strings.ToUpper(direction), name)
			}
		}
	}
	if up != "" || down != "" || in != "" || out != "" {
		s.WriteLine("")
		s.WriteLine("\x1b[1;37mOTHER PATHS\x1b[0m")
		if up != "" { s.WriteLine("  UP    → %s", up) }
		if down != "" { s.WriteLine("  DOWN  → %s", down) }
		if in != "" { s.WriteLine("  IN    → %s", in) }
		if out != "" { s.WriteLine("  OUT   → %s", out) }
	}
	if len(r.Exits) == 0 { s.WriteLine("  No exits are visible.") }
	s.WriteLine("")
	s.WriteLine("\x1b[90mUse 'exits' for the full list or 'worldmap' for the wider world.\x1b[0m")
}
func (s *Session) look() {
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil {
		s.WriteLine("\x1b[1;31mYou are nowhere.\x1b[0m The world has lost track of you.")
		return
	}

	s.WriteLine("")
	s.WriteLine("\x1b[1;36m╭─ %s ─╮\x1b[0m", r.Name)
	s.WriteLine("%s", compactText(r.Description, 125))

	directions := []string{"north", "south", "east", "west", "up", "down", "in", "out"}
	var visible []string
	for _, d := range directions {
		if id, ok := r.Exits[d]; ok {
			if dest := s.World.GetRoom(id); dest != nil {
				visible = append(visible, fmt.Sprintf("%-5s → %s", strings.ToUpper(d), dest.Name))
			}
		}
	}
	if len(visible) > 0 {
		s.WriteLine("")
		s.WriteLine("\x1b[1;37mPATHS\x1b[0m")
		for _, exit := range visible {
			s.WriteLine("  \x1b[36m›\x1b[0m %s", exit)
		}
	}

	npcs := s.World.NPCs[s.Character.RoomID]
	if len(npcs) > 0 {
		s.WriteLine("")
		s.WriteLine("\x1b[1;32mPEOPLE TO TALK TO\x1b[0m")
		for i, n := range npcs {
			s.WriteLine("  \x1b[32m[NPC %d] %s\x1b[0m", i+1, n.Name)
			if n.Description != "" {
				s.WriteLine("         %s", compactText(n.Description, 88))
			}
		}
		s.WriteLine("  \x1b[32mTalk:\x1b[0m talk <number> <topic>   \x1b[32mShop:\x1b[0m shop <number>")
	}

	if s.Enemy == nil && s.EncounterClearedRoom != s.Character.RoomID {
		s.spawnEnemy()
	}
	if s.Enemy != nil && s.Enemy.HP > 0 {
		s.WriteLine("")
		s.WriteLine("\x1b[1;33mTHREATS\x1b[0m")
		s.WriteLine("  \x1b[1;33m[ENEMY 1] %s\x1b[0m", s.Enemy.Name)
		if s.Enemy.Description != "" {
			s.WriteLine("           %s", compactText(s.Enemy.Description, 96))
		}
		s.WriteLine("           Health: %d / %d", s.Enemy.HP, s.Enemy.MaxHP)
		s.WriteLine("  \x1b[33mFight:\x1b[0m attack 1  (or attack %s)", strings.ToLower(s.Enemy.Name))
	} else if s.EncounterClearedRoom == s.Character.RoomID {
		s.WriteLine("")
		s.WriteLine("\x1b[1;32mTHREATS\x1b[0m")
		s.WriteLine("  The area is quiet. The defeated creature has not returned.")
	}

}
func compactText(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	if limit < 4 {
		return string(runes[:limit])
	}
	return string(runes[:limit-1]) + "…"
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
	case "cerberus_gate":
		s.Enemy = &combat.Enemy{Name: "Cerberus", Description: "Three enormous heads rise from the darkness. One snarls, one watches, and one whispers your name.", Level: 9, HP: 350, MaxHP: 350, Defense: 12, Damage: 28, XP: 450}
	case "tartarus_edge":
		s.Enemy = &combat.Enemy{Name: "Tartarus Brute", Description: "A colossal shape claws its way up from the abyss, dragging chains that were forged before the gods had names.", Level: 10, HP: 390, MaxHP: 390, Defense: 13, Damage: 31, XP: 520}
	case "modern_plaka":
		s.Enemy = &combat.Enemy{Name: "Street Shade", Description: "A human-shaped shadow slips between the neon signs, moving against the light.", Level: 2, HP: 80, MaxHP: 80, Defense: 5, Damage: 10, XP: 90}
	case "modern_metro":
		s.Enemy = &combat.Enemy{Name: "Echo Hound", Description: "A hound made of static and old memories emerges from the empty train.", Level: 4, HP: 145, MaxHP: 145, Defense: 7, Damage: 15, XP: 170}
	case "modern_styx":
		s.Enemy = &combat.Enemy{Name: "Styx Wraith", Description: "Black water rises into the shape of a veiled figure, carrying coins from lives you never lived.", Level: 5, HP: 185, MaxHP: 185, Defense: 8, Damage: 18, XP: 230}
	case "modern_sanctum":
		s.Enemy = &combat.Enemy{Name: "Bronze Sentinel", Description: "A bronze guardian unfolds from the hidden temple wall, its eyes burning with borrowed starlight.", Level: 6, HP: 220, MaxHP: 220, Defense: 9, Damage: 20, XP: 280}
	case "future_city":
		s.Enemy = &combat.Enemy{Name: "Chronal Warden", Description: "A guardian of fractured seconds steps from a ripple in the air.", Level: 2, HP: 95, MaxHP: 95, Defense: 5, Damage: 11, XP: 100}
	case "future_skyway":
		s.Enemy = &combat.Enemy{Name: "Storm Automaton", Description: "A machine of celestial bronze and lightning blocks the road above the clouds.", Level: 4, HP: 150, MaxHP: 150, Defense: 7, Damage: 15, XP: 180}
	case "future_moon":
		s.Enemy = &combat.Enemy{Name: "Moonshade", Description: "A pale shadow detaches itself from the lunar sanctuary's wall and reaches for your memories.", Level: 4, HP: 145, MaxHP: 145, Defense: 7, Damage: 14, XP: 180}
	case "far_era":
		s.Enemy = &combat.Enemy{Name: "Last Shore Titan", Description: "The drowned world's final guardian rises from the surf, carrying the weight of vanished centuries.", Level: 5, HP: 180, MaxHP: 180, Defense: 8, Damage: 18, XP: 220}
	default:
		return
	}
	s.WriteLine("\x1b[1;31mA %s appears!\x1b[0m",s.Enemy.Name)
	s.WriteLine("%s",s.Enemy.Description)
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
	if r.ID == "village_lane" && direction == "north" && s.TutorialStep < 4 {
		s.WriteLine("Damon raises a hand. Finish the village lessons first: look, talk to the guide, use hint, then ask about the road.")
		return
	}
	s.Character.RoomID=next
	s.EncounterClearedRoom = ""
	s.PendingEquip = nil
	s.recordDiscovery(next)
	s.look()
	if s.AutoMap { s.localMap() }
	s.updateQuests()
	if s.Character.RoomID == "village_lane" && s.TutorialStep == 0 {
		s.advanceTutorial(1)
	}
}

func (s *Session) attack(args []string) {
	if s.Enemy == nil || s.Enemy.HP <= 0 {
		s.WriteLine("There is nothing here to fight.")
		return
	}
	if len(args)>0{target:=strings.Join(args," ");if target!="1"&&!strings.Contains(strings.ToLower(s.Enemy.Name),target){s.WriteLine("Your current target is %s. Use 'attack 1' or 'attack %s'.",s.Enemy.Name,strings.ToLower(s.Enemy.Name));return}}
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
	s.EncounterClearedRoom = s.Character.RoomID
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
	if len(args) == 0 {
		s.WriteLine("Equip what? Try 'eq' to inspect your gear, or 'inv' to list items.")
		return
	}

	// A numbered choice resolves the ambiguity from the most recent partial match.
	if len(args) == 1 {
		if n, err := strconv.Atoi(args[0]); err == nil {
			if n < 1 || n > len(s.PendingEquip) {
				s.WriteLine("That selection is no longer available. Try 'wield <part of item name>' again.")
				s.PendingEquip = nil
				return
			}
			name := s.PendingEquip[n-1]
			s.PendingEquip = nil
			s.equip([]string{name})
			return
		}
	}

	query := strings.ToLower(strings.Join(args, " "))
	var exact []item.Item
	var matches []item.Item
	for _, candidate := range s.Character.Inventory {
		name := strings.ToLower(candidate.Name)
		if name == query {
			exact = append(exact, candidate)
			continue
		}
		if strings.Contains(name, query) {
			matches = append(matches, candidate)
		}
	}

	if len(exact) == 1 {
		s.equipItem(exact[0])
		return
	}
	if len(exact) == 0 && len(matches) == 1 {
		s.equipItem(matches[0])
		return
	}
	if len(exact) > 1 {
		matches = exact
	}
	if len(matches) > 1 {
		s.PendingEquip = make([]string, len(matches))
		s.WriteLine("Several items match '%s'. Type 'equip <number>' to choose:", query)
		for i, candidate := range matches {
			s.PendingEquip[i] = candidate.Name
			s.WriteLine("  %d) %s", i+1, candidate.Name)
		}
		return
	}
	s.PendingEquip = nil
	s.WriteLine("No carried equipment matches '%s'. Try 'inv' to see your items.", query)
}

func (s *Session) equipItem(i item.Item) {
	if item.IsWeapon(i) {
		s.Character.Weapon = i.Name
		s.WriteLine("You wield %s.", i.Name)
		return
	}
	if i.Kind == "armor" {
		s.Character.Armor = i.Name
		s.WriteLine("You wear %s. Armor: %d.", i.Name, i.Armor)
		return
	}
	s.WriteLine("%s cannot be equipped.", i.Name)
}

func (s *Session) inventory() {
	s.WriteLine("")
	s.WriteLine("\x1b[1;36m╭─ INVENTORY ─╮\x1b[0m")
	if len(s.Character.Inventory) == 0 {
		s.WriteLine("  Your pack is empty.")
		return
	}
	s.WriteLine("  Items carried: %d", len(s.Character.Inventory))
	for i, it := range s.Character.Inventory {
		tags := ""
		if strings.EqualFold(it.Name, s.Character.Weapon) || strings.EqualFold(it.Name, s.Character.Armor) {
			tags += "  \x1b[1;32m[EQUIPPED]\x1b[0m"
		}
		if it.Relic {
			tags += "  \x1b[1;35m[SOUL RELIC]\x1b[0m"
		}
		s.WriteLine("  %2d. %s%s%s%s", i+1, s.color(it.TierColor()), it.Name, s.color("0"), tags)
		s.WriteLine("      %s  |  Armor +%d  |  Damage +%d", it.TierName(), it.Armor, it.Damage)
	}
	s.WriteLine("")
	s.WriteLine("  Equip: wield <name>  |  Armor: wear <name>")
}
func (s *Session) equipment() {
	s.WriteLine("")
	s.WriteLine("\x1b[1;36m╭─ EQUIPPED GEAR ─╮\x1b[0m")
	weapon := s.Character.Weapon
	armor := s.Character.Armor
	if weapon == "" { weapon = "None" }
	if armor == "" { armor = "None" }
	s.WriteLine("  Weapon  ⚔  %s", weapon)
	s.WriteLine("  Armor   ◈  %s", armor)
	s.WriteLine("")
	s.WriteLine("  Change weapon: wield <item name>")
	s.WriteLine("  Change armor:  wear <item name>")
	s.WriteLine("  Tip: short names work, e.g. wield heph.")
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
	c := s.Character
	s.WriteLine("")
	s.WriteLine("\x1b[1;36m╭─ %s ─╮\x1b[0m", c.Name)
	s.WriteLine("\x1b[1;37mIDENTITY\x1b[0m")
	s.WriteLine("  Level %d   •   Life %d   •   Rebirths %d", c.Level, c.Life, c.Rebirths)
	s.WriteLine("  Era: %s   |   Domain: %s", c.Era, c.Domain)
	s.WriteLine("")
	s.WriteLine("\x1b[1;37mVITALS\x1b[0m")
	s.WriteLine("  HP    %d / %d", c.HP, c.MaxHP)
	s.WriteLine("  Mana  %d / %d", c.Mana, c.MaxMana)
	s.WriteLine("  XP    %d   |   Next level in %d", c.Experience, progression.XPToNextLevel(c.Level, c.Experience))
	s.WriteLine("")
	s.WriteLine("\x1b[1;37mCOMBAT & RESOURCES\x1b[0m")
	s.WriteLine("  Attack %d   |   Defense %d   |   Armor %d", c.AttackPower(), c.DefensePower(), c.ArmorPower())
	s.WriteLine("  Divinity %d   |   Gold %d", c.Divinity, c.Gold)
	s.WriteLine("")
	s.WriteLine("\x1b[1;37mSOUL LEGACY\x1b[0m")
	s.WriteLine("  Memories %d  •  Scars %d  •  Oaths %d", len(c.Memories), len(c.Scars), len(c.Oaths))
	s.WriteLine("  Favors %d  •  Curses %d  •  Echoes %d", len(c.Favors), len(c.Curses), len(c.Echoes))
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
	if s.Enemy != nil && s.Enemy.HP > 0 {
		s.WriteLine("The battle holds you to this life. Defeat the enemy or flee before rebirth.")
		return
	}
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

func (s *Session) currentNPC(index int) *world.NPC { npcs:=s.World.NPCs[s.Character.RoomID];if index<1||index>len(npcs){return nil};return npcs[index-1] }
func(s *Session) parseNPCSelection(args []string)(int,[]string){if len(args)==0{return 1,args};if n,err:=strconv.Atoi(args[0]);err==nil{return n,args[1:]};return 1,args}

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
 if s.Enemy!=nil&&s.Enemy.HP>0{s.WriteLine("The %s keeps you too busy to talk.",s.Enemy.Name);return}
 index,topicArgs:=s.parseNPCSelection(args);n:=s.currentNPC(index)
 if n==nil{if len(s.World.NPCs[s.Character.RoomID])==0{s.WriteLine("There is no one here willing to speak with you.")}else{s.WriteLine("Choose a valid NPC number from 'look'.")};return}
 topic:="hello";if len(topicArgs)>0{topic=strings.Join(topicArgs," ")}
 if greeting:=s.persistentNPCGreeting(n,topic);greeting!=""{s.WriteLine("%s",greeting);return}
 if topic=="hello"||topic=="greeting"{if greeting:=s.factionGreeting(n);greeting!=""{s.WriteLine("%s",greeting);return}}
 if n.ID=="pythia"&&(topic=="hello"||topic=="greeting"){switch{case s.Character.HasStoryFlag("oracle_trust"):s.WriteLine("\x1b[1;36mPythia smiles faintly. \"You kept the thread I gave you. The river will test that promise when you least expect it.\"\x1b[0m");return;case s.Character.HasStoryFlag("oracle_defied"):s.WriteLine("\x1b[1;36mPythia regards you without anger. \"Still walking your own road, I see. Even the Fates have learned to leave a little room for defiance.\"\x1b[0m");return}}
 if topic=="choice"||topic=="choices"{if n.ID=="pythia"{if s.Character.HasStoryFlag("oracle_choice_made"){s.WriteLine("Pythia studies you. \"The river has recorded your answer. You cannot make that choice unmade.\"");return};s.WriteLine("Pythia's voice falls to a whisper: \"When the Fates offer a thread, will you trust the pattern or cut your own path?\"");s.WriteLine("  choose trust — accept the Oracle's guidance and swear to remember it.");s.WriteLine("  choose defy  — reject prophecy and bear the consequences alone.");return}}
 if t:=n.DialogueFor(topic);t!="" {
		s.WriteLine("\x1b[1;36m%s:\x1b[0m %s",n.Name,t)
		if n.ID == "village_guide" && s.TutorialStep == 2 && (topic == "hello" || topic == "greeting") {
			s.advanceTutorial(3)
		} else if n.ID == "village_guide" && s.TutorialStep == 3 && (topic == "road" || topic == "mountain" || topic == "north") {
			s.advanceTutorial(4)
		}
		return
	}
	s.WriteLine("\x1b[1;36m%s:\x1b[0m \"Ask me about the things that matter here.\"",n.Name)
}

func (s *Session) persistentNPCGreeting(n *world.NPC, topic string) string {
	if n == nil || n.ID != "athens_vendor" {
		return ""
	}
	if (topic == "museum" || topic == "artifact" || topic == "replica") && s.Character.HasStoryFlag("museum_echo_uncovered") {
		return "Myrto glances toward the museum district. \"Some things behind glass are not exhibits. If one remembered you, keep that memory close.\""
	}
	if topic != "hello" && topic != "greeting" {
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
	if s.Enemy != nil && s.Enemy.HP > 0 {
		s.WriteLine("The %s leaves no room for a choice right now.", s.Enemy.Name)
		return
	}
	if len(args) == 0 {
		s.WriteLine("Choose what? At the Oracle, try 'talk choices' first.")
		return
	}
	n := s.currentNPC(1)
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

func(s *Session) shop(args []string) {
 index,_:=s.parseNPCSelection(args);n:=s.currentNPC(index);if n==nil||!n.HasShop(){s.WriteLine("There is no merchant at that NPC number.");return}
 s.WriteLine("\x1b[1;33m%s's WARES\x1b[0m — You have %d gold",n.Name,s.Character.Gold)
 for i,ware:=range n.Shop{price:=s.merchantPrice(n,ware);note:="";if price<ware.Price{note=" (faction discount)"}else if price>ware.Price{note=" (faction surcharge)"};s.WriteLine("  %d) %s%s %s — %d gold%s",i+1,s.color(ware.TierColor()),ware.TierName(),ware.Name,price,note)}
 s.WriteLine("Use 'buy <number>' or 'buy <part of item name>'.")
}

func(s *Session) buy(args []string) {
 if len(args)==0{s.WriteLine("Buy what? Try 'shop'.");return};n:=s.currentNPC(1);if n==nil||!n.HasShop(){s.WriteLine("There is no merchant here.");return}
 var selected *item.Item
 if len(args)==1{if number,err:=strconv.Atoi(args[0]);err==nil{if number<1||number>len(n.Shop){s.WriteLine("Choose an item number shown by 'shop'.");return};selected=&n.Shop[number-1]}}
 if selected==nil{query:=strings.ToLower(strings.Join(args," "));var matches []int;for i:=range n.Shop{if strings.EqualFold(n.Shop[i].Name,query){selected=&n.Shop[i];break};if strings.Contains(strings.ToLower(n.Shop[i].Name),query){matches=append(matches,i)}}
 if selected==nil&&len(matches)==1{selected=&n.Shop[matches[0]]};if selected==nil&&len(matches)>1{s.WriteLine("Several wares match '%s'. Choose a number:",query);for _,i:=range matches{s.WriteLine("  %d) %s",i+1,n.Shop[i].Name)};return}}
 if selected==nil{s.WriteLine("That item is not for sale here. Type 'shop' to see numbered wares.");return};price:=s.merchantPrice(n,*selected);if s.Character.Gold<price{s.WriteLine("You need %d more gold.",price-s.Character.Gold);return}
 for _,owned:=range s.Character.Inventory{if strings.EqualFold(owned.Name,selected.Name){s.WriteLine("You already possess that item.");return}}
 s.Character.Gold-=price;s.Character.Inventory=append(s.Character.Inventory,*selected);s.WriteLine("\x1b[1;32mPurchased %s for %d gold.\x1b[0m",selected.Name,price)
}

