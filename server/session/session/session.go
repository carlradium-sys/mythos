package session

import (
	"bufio"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

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

type onlineSnapshot struct {
	Name string
	Level int
	Life int
	Era string
	RoomID string
}

var onlineMu sync.RWMutex
var onlineSessions = map[*Session]onlineSnapshot{}

type Session struct {
	writeMu sync.Mutex
 Character *character.Character
 Account *account.Account
 Accounts *account.Store
 Conn net.Conn
 World *world.World
 Enemy *combat.Enemy
 EnemyAggressive bool
 Guarding bool
 PrologueStep int
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

// WriteLine centralizes Telnet line endings and wraps long prose without
// counting ANSI color codes as visible terminal columns.
func (s *Session) WriteLine(format string, args ...any) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	rendered := fmt.Sprintf(format, args...)
	rendered = strings.ReplaceAll(rendered, "\r\n", "\n")
	for _, line := range strings.Split(rendered, "\n") {
		for _, wrapped := range wrapTerminalLine(line, 78) {
			fmt.Fprint(s.Conn, wrapped, "\r\n")
		}
	}
}

func visibleTerminalWidth(value string) int {
	width := 0
	for i := 0; i < len(value); {
		if value[i] == 0x1b && i+1 < len(value) && value[i+1] == '[' {
			i += 2
			for i < len(value) {
				ch := value[i]
				i++
				if ch >= 0x40 && ch <= 0x7e { break }
			}
			continue
		}
		_, size := utf8.DecodeRuneInString(value[i:])
		if size < 1 { size = 1 }
		width++
		i += size
	}
	return width
}

func wrapTerminalLine(line string, width int) []string {
	if width < 20 || visibleTerminalWidth(line) <= width || strings.ContainsAny(line, "╭╮╰╯│┌┐└┘←→↑↓") {
		return []string{line}
	}
	words := strings.Fields(line)
	if len(words) < 2 { return []string{line} }
	indent := len(line) - len(strings.TrimLeft(line, " "))
	prefix := strings.Repeat(" ", indent)
	var lines []string
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		if visibleTerminalWidth(current)+1+visibleTerminalWidth(word) > width {
			lines = append(lines, current)
			current = prefix + word
		} else {
			current += " " + word
		}
	}
	if current != "" { lines = append(lines, current) }
	return lines
}

func (s *Session) Run(scanner *bufio.Scanner) {
	s.WriteLine("")
	s.WriteLine("\x1b[1;36mFATEWALKER: BEYOND THE STYX\x1b[0m")
	s.WriteLine("\x1b[90mA Greek mythic fantasy of death, choice, and rebirth.\x1b[0m")
	if !s.login(scanner) { return }
	defer s.persist()
	if !s.selectCharacter(scanner) { return }
	s.registerOnline()
	defer s.unregisterOnline()
	s.prepareTutorial()
	s.storyIntro()
	s.updateQuests()
	s.look()
	s.refreshOnline()
	if s.Character.HasStoryFlag("tutorial_available") {
		s.prologueHint()
	} else if s.TutorialStep < 16 {
		s.tutorialHint()
	}
	// Read input asynchronously so the combat ticker can keep fighting while
	// the player is deciding what to do, just like a traditional MUD.
	commands := make(chan string)
	done := make(chan struct{})
	go func() {
		defer close(commands)
		for scanner.Scan() {
			select {
			case commands <- scanner.Text():
			case <-done:
				return
			}
		}
	}()
	defer close(done)
	combatTicker := time.NewTicker(2500 * time.Millisecond)
	defer combatTicker.Stop()
	s.WriteLine("")
	s.WriteLine("%s", s.prompt())
	for {
		select {
		case line, ok := <-commands:
			if !ok { return }
			if !s.handleCommand(strings.TrimSpace(line)) { return }
			s.refreshOnline()
			s.WriteLine("")
			s.WriteLine("%s", s.prompt())
		case <-combatTicker.C:
			if s.Character != nil && s.Character.AutoCombat && s.Enemy != nil && s.Enemy.HP > 0 && s.EnemyAggressive {
				s.automaticCombatRound()
				s.refreshOnline()
				s.persist()
				s.WriteLine("%s", s.prompt())
			}
		}
	}
}

func (s *Session) prompt() string {
	return fmt.Sprintf("\x1b[1;36m[%s L%d Life %d HP %d/%d]\x1b[0m > ",s.Character.Era,s.Character.Level,s.Character.Life,s.Character.HP,s.Character.MaxHP)
}

func (s *Session) handleCommand(input string) bool {
	rawParts := strings.Fields(input)
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
	case "look", "l":
		if len(parts) > 1 {
			s.lookTarget(parts[1:])
			if s.TutorialStep == 2 && s.Character.HasStoryFlag("tutorial_inspected_npc") { s.advanceTutorial(3) }
			if s.TutorialStep == 10 && s.Character.HasStoryFlag("tutorial_enemy_inspected") { s.advanceTutorial(11) }
		} else {
			s.look()
			if s.TutorialStep == 0 { s.advanceTutorial(1) }
		}
	case "examine", "inspect":
		s.examineTarget(parts[1:])
	case "listen":
		s.listen(parts[1:])
	case "get", "take", "pickup":
		s.getObject(parts[1:])
	case "drop", "discard":
		s.dropItem(parts[1:])
	case "put", "place":
		s.putItem(parts[1:])
	case "open":
		s.openContainer(parts[1:])
	case "close":
		s.closeContainer(parts[1:])
	case "read":
		s.readObject(parts[1:])
	case "search":
		s.searchObject(parts[1:])
	case "scan":
		s.scanTargets()
	case "iscan", "itemscan", "scanitems":
		s.scanItems()
	case "map":
		s.localMap()
		if s.TutorialStep == 8 { s.advanceTutorial(9) }
	case "automap", "mapauto", "maptoggle":
		if parts[0] == "maptoggle" && len(parts) == 1 {
			s.autoMap([]string{"toggle"})
		} else {
			s.autoMap(parts[1:])
		}
	case "worldmap", "world-map":
		s.WriteLine("%s", s.World.MapText(s.Character.RoomID))
	case "exits":
		s.showExits()
		if s.TutorialStep == 7 { s.advanceTutorial(8) }
	case "north","south","east","west","up","down","in","out","n","s","e","w","u","d":
		s.move(parts[0])
	case "who":
		s.who()
	case "chat", "gossip":
		s.chat(rawParts[1:])
	case "emote", "me":
		s.emote(rawParts[1:])
	case "tell", "whisper":
		s.tell(rawParts[1:])
	case "score","stats":
		s.score()
	case "inventory", "inv", "i":
		s.inventory()
		if s.TutorialStep == 12 { s.advanceTutorial(13) }
	case "equipment", "eq", "gear":
		s.equipment()
		if s.TutorialStep == 13 { s.advanceTutorial(14) }
	case "equip", "wield", "wear":
		s.equip(parts[1:])
	case "talk":
		s.talk(parts[1:])
	case "say":
		if s.isNPCAddress(parts[1:]) { s.talk(parts[1:]) } else { s.sayRoom(rawParts[1:]) }
	case "hint":
		s.hint(parts[1:])
		if s.TutorialStep == 5 && s.Character.HasStoryFlag("tutorial_hint_used") { s.advanceTutorial(6) }
	case "choose":
		s.choose(parts[1:])
	case "offer", "give":
		s.offer(parts[1:])
	case "shop", "wares":
		s.shop(parts[1:])
	case "buy":
		s.buy(parts[1:])
	case "craft", "forge":
		s.craft(parts[1:])
	case "use", "drink", "consume":
		s.useItem(parts[1:])
	case "attack","kill","hit":
		s.attack(parts[1:])
	case "pc":
		s.combatMode(parts[1:], true)
	case "combatmode", "combat-mode":
		s.combatMode(parts[1:], false)
	case "cast":
		s.cast(parts[1:])
	case "ql", "quicklist":
		s.quicklist(parts[1:])
	case "powers","power":
		s.powers()
	case "awaken":
		s.awaken(parts[1:])
	case "flee":
		s.flee()
	case "guard", "defend":
		s.guard()
	case "rest":
		s.rest()
	case "rebirth":
		s.rebirth(parts[1:])
	case "journal","quest","story":
		s.journal()
		if s.TutorialStep == 15 { s.advanceTutorial(16) }
	case "quests":
		s.questList()
		if s.TutorialStep == 14 { s.advanceTutorial(15) }
	case "current":
		s.currentQuests(parts[1:])
	case "soul","legacy":
		s.soul()
	case "invoke","gift":
		s.invokeGift()
	default:
		s.WriteLine("Unknown command. Type 'help' for help.")
	}
	if s.Character.HasStoryFlag("tutorial_available") { s.advancePrologueForCommand(parts) }
	s.persist()
	return true
}

func (s *Session) registerOnline() {
	if s.Character == nil { return }
	onlineMu.Lock()
	onlineSessions[s] = onlineSnapshot{Name:s.Character.Name, Level:s.Character.Level, Life:s.Character.Life, Era:s.Character.Era, RoomID:s.Character.RoomID}
	onlineMu.Unlock()
	s.broadcastGlobalExcept(fmt.Sprintf("%s enters the world.", s.Character.Name), s)
}

func (s *Session) refreshOnline() {
	if s.Character == nil { return }
	onlineMu.Lock()
	if _, ok := onlineSessions[s]; ok {
		onlineSessions[s] = onlineSnapshot{Name:s.Character.Name, Level:s.Character.Level, Life:s.Character.Life, Era:s.Character.Era, RoomID:s.Character.RoomID}
	}
	onlineMu.Unlock()
}

func (s *Session) unregisterOnline() {
	onlineMu.Lock()
	snapshot, ok := onlineSessions[s]
	if ok { delete(onlineSessions, s) }
	onlineMu.Unlock()
	if ok { s.broadcastGlobal(fmt.Sprintf("%s leaves the world.", snapshot.Name)) }
}

func (s *Session) onlineSnapshotList() []onlineSnapshot {
	onlineMu.RLock()
	list := make([]onlineSnapshot, 0, len(onlineSessions))
	for _, player := range onlineSessions { list = append(list, player) }
	onlineMu.RUnlock()
	sort.Slice(list, func(i,j int) bool { return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name) })
	return list
}

func (s *Session) who() {
	players := s.onlineSnapshotList()
	s.WriteLine("\x1b[1;36mPLAYERS ONLINE — %d\x1b[0m", len(players))
	if len(players) == 0 { s.WriteLine("  No travelers are visible."); return }
	for _, player := range players {
		room := player.RoomID
		if found := s.World.GetRoom(room); found != nil { room = found.Name }
		s.WriteLine("  %-20s Level %-3d Life %-2d %-8s — %s", player.Name, player.Level, player.Life, player.Era, room)
	}
}

func (s *Session) broadcastGlobal(message string) {
	s.broadcastGlobalExcept(message, nil)
}

func (s *Session) broadcastGlobalExcept(message string, except *Session) {
	onlineMu.RLock()
	recipients := make([]*Session, 0, len(onlineSessions))
	for player := range onlineSessions { if player != except { recipients = append(recipients, player) } }
	onlineMu.RUnlock()
	for _, player := range recipients { player.WriteLine("\x1b[1;35m[World]\x1b[0m %s", message) }
}

func (s *Session) broadcastRoom(message string) {
	s.refreshOnline()
	onlineMu.RLock()
	recipients := make([]*Session, 0, len(onlineSessions))
	for player, snapshot := range onlineSessions {
		if snapshot.RoomID == s.Character.RoomID { recipients = append(recipients, player) }
	}
	onlineMu.RUnlock()
	for _, player := range recipients { player.WriteLine("%s", message) }
}

func (s *Session) chat(args []string) {
	message := strings.TrimSpace(strings.Join(args, " "))
	if message == "" { s.WriteLine("Usage: chat <message>"); return }
	s.broadcastGlobal(fmt.Sprintf("%s: %s", s.Character.Name, message))
}

func (s *Session) emote(args []string) {
	message := strings.TrimSpace(strings.Join(args, " "))
	if message == "" { s.WriteLine("Usage: emote <action>"); return }
	s.broadcastRoom(fmt.Sprintf("\x1b[1;35m%s %s\x1b[0m", s.Character.Name, message))
}

func (s *Session) sayRoom(args []string) {
	message := strings.TrimSpace(strings.Join(args, " "))
	if message == "" { s.WriteLine("Say what?"); return }
	s.broadcastRoom(fmt.Sprintf("\x1b[1;36m%s says, %q\x1b[0m", s.Character.Name, message))
}

func (s *Session) isNPCAddress(args []string) bool {
	if len(args) == 0 { return false }
	first := strings.ToLower(args[0])
	if _, err := strconv.Atoi(first); err == nil { return len(s.World.NPCs[s.Character.RoomID]) > 0 }
	if len(first) < 3 { return false }
	for _, n := range s.World.NPCs[s.Character.RoomID] {
		if strings.Contains(strings.ToLower(n.Name), first) { return true }
	}
	return false
}

func (s *Session) tell(args []string) {
	if len(args) < 2 { s.WriteLine("Usage: tell <player> <message>"); return }
	targetName := strings.ToLower(args[0])
	message := strings.TrimSpace(strings.Join(args[1:], " "))
	onlineMu.RLock()
	var target *Session
	var name string
	var partialMatches []*Session
	var partialNames []string
	for player, snapshot := range onlineSessions {
		if strings.EqualFold(snapshot.Name, targetName) {
			target, name = player, snapshot.Name
			break
		}
		if strings.Contains(strings.ToLower(snapshot.Name), targetName) {
			partialMatches = append(partialMatches, player)
			partialNames = append(partialNames, snapshot.Name)
		}
	}
	onlineMu.RUnlock()
	if target == nil && len(partialMatches) == 1 { target, name = partialMatches[0], partialNames[0] }
	if target == nil && len(partialMatches) > 1 {
		s.WriteLine("That name matches several players: %s. Use a more specific name.", strings.Join(partialNames, ", "))
		return
	}
	if target == nil { s.WriteLine("No online player matches '%s'. Use 'who' to see who's here.", args[0]); return }
	if target == s { s.WriteLine("You tell yourself: %s", message); return }
	target.WriteLine("\x1b[1;35m[Tell from %s]\x1b[0m %s", s.Character.Name, message)
	s.WriteLine("\x1b[1;35m[Tell to %s]\x1b[0m %s", name, message)
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
   s.Account.Characters=append(s.Account.Characters,*rec);s.persist();s.Character=rec.Character;s.initializeCombatMode();s.WriteLine("%s created.",s.Character.Name);return true
  case "delete":
   if len(p)!=2{s.WriteLine("Delete which character number?");continue}
   n,err:=strconv.Atoi(p[1]);if err!=nil||n<1||n>len(s.Account.Characters){s.WriteLine("Invalid character number.");continue}
   s.Account.Characters=append(s.Account.Characters[:n-1],s.Account.Characters[n:]...);s.persist();s.WriteLine("Character deleted.")
  default:
   n,err:=strconv.Atoi(p[0]);if err!=nil||n<1||n>len(s.Account.Characters){s.WriteLine("Choose a character number.");continue}
   s.Character=s.Account.Characters[n-1].Character;s.Character.EnsureLifeGift();s.initializeCombatMode();s.WriteLine("Welcome back, %s.",s.Character.Name);return true
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
	if q.RewardGold > 0 {
		s.Character.Gold += q.RewardGold
		s.WriteLine("You receive %d gold for completing %s.", q.RewardGold, q.Name)
	}
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
	case "prayer_without_name":
		s.Character.SetStoryFlag("prayer_witnessed")
		s.Character.Memories = append(s.Character.Memories, "The nameless prayer asked that someone remember why a person stayed, not who they were.")
		s.Character.Echoes = append(s.Character.Echoes, "A witness mark survives where a name was erased.")
		s.WriteLine("The black thread tightens once around your wrist, then relaxes. You have not learned the writer's name, but you understand what they refused to let the world forget.")
	case "return_the_coin":
		s.Character.SetStoryFlag("well_answered")
		s.Character.Echoes = append(s.Character.Echoes, "The village well remembered the shore where the scratched coin first changed hands.")
		s.WriteLine("The well has answered, though it has not explained itself. Your choice is now part of the thread.")
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
  goalReached := q.TargetRoom!="" && q.TargetEnemy=="" && s.Character.RoomID==q.TargetRoom
  if q.CompletionFlag!="" && s.Character.HasStoryFlag(q.CompletionFlag) { goalReached = true }
  if goalReached {
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
func (s *Session) currentQuests(args []string) {
	s.ensureQuests()
	if len(args) > 0 && args[0] == "track" {
		if len(args) != 2 { s.WriteLine("Usage: current quests track <ID>"); return }
		id, err := strconv.Atoi(args[1])
		q, found := quest.GetByNumber(id)
		if err != nil || !found { s.WriteLine("Invalid quest ID. Use 'current quests' to see IDs."); return }
		if !s.questUnlocked(q) || s.Character.Quests[q.ID] >= q.Required { s.WriteLine("That quest is not active. Choose one from 'current quests'."); return }
		s.Character.TrackedQuest = q.ID
		s.WriteLine("Quest tracker set to [%d] %s.", id, q.Name)
		s.showCurrentQuest(q); return
	}
	if len(args) > 0 && args[0] == "untrack" { s.Character.TrackedQuest = ""; s.WriteLine("Quest tracker cleared."); return }
	if len(args) == 1 && args[0] == "quest" {
		if q, ok := quest.Get(s.Character.TrackedQuest); ok { s.showCurrentQuest(q); return }
		if q, ok := s.nextAvailableQuest(); ok { s.Character.TrackedQuest = q.ID; s.showCurrentQuest(q); return }
		s.WriteLine("You have no active quest to display."); return
	}
	if len(args) == 0 || (len(args) == 1 && args[0] == "quests") { s.currentQuestList(); return }
	s.WriteLine("Usage: current quests | current quests track <ID> | current quest | current quests untrack")
}

func (s *Session) currentQuestList() {
	active := 0
	s.WriteLine("\x1b[1;36mCURRENT QUESTS\x1b[0m")
	for _, q := range quest.All() {
		p := s.Character.Quests[q.ID]
		if !s.questUnlocked(q) || p >= q.Required { continue }
		active++; marker := " "
		if s.Character.TrackedQuest == q.ID { marker = "*" }
		s.WriteLine(" %s[%d] %-28s (%d/%d)", marker, q.Number, q.Name, p, q.Required)
		s.WriteLine("     %s", q.Goal)
	}
	if active == 0 { s.WriteLine("You have no active quests at the moment. Explore, speak with people, and examine anything that seems out of place."); return }
	s.WriteLine("Use 'current quests track <ID>' to pin a quest to your tracker.")
	if q, ok := quest.Get(s.Character.TrackedQuest); ok { s.WriteLine("Tracked: %s. Type 'current quest' for details.", q.Name) }
}

func (s *Session) showCurrentQuest(q quest.Quest) {
	p := s.Character.Quests[q.ID]; if p > q.Required { p = q.Required }
	status := "In progress"; if p >= q.Required { status = "Complete" }
	s.WriteLine(""); s.WriteLine("\x1b[1;36mCURRENT QUEST — %s\x1b[0m", q.Name)
	s.WriteLine("Quest ID: %s", q.ID); s.WriteLine("Status: %s (%d/%d)", status, p, q.Required)
	s.WriteLine(""); s.WriteLine("%s", q.Description); s.WriteLine(""); s.WriteLine("Your task: %s", q.Goal)
	s.WriteLine("Progress: %d of %d objective(s) complete.", p, q.Required)
	steps := questTrackerSteps(q.ID)
	if len(steps) > 0 {
		s.WriteLine("")
		s.WriteLine("What you have uncovered:")
		for _, step := range steps {
			done := s.Character.HasStoryFlag(step.Flag)
			if p >= q.Required { done = true }
			mark := "[ ]"
			if done { mark = "[x]" }
			s.WriteLine("  %s %s", mark, step.Text)
		}
	}
	if p < q.Required {
		if hint := s.questNextLead(q, steps); hint != "" {
			s.WriteLine("Next lead: %s", hint)
		}
	}
	if q.RewardXP > 0 { s.WriteLine("Reward: %d XP and %d gold.", q.RewardXP, q.RewardGold) }
	s.WriteLine(""); s.WriteLine("Tip: use 'look', 'examine <thing>', 'listen', or 'talk <person>' when following a clue.")
}

type questTrackerStep struct {
	Flag string
	Text string
	Hint string
}

// questTrackerSteps turns persistent discoveries into a concise, story-aware checklist.
func questTrackerSteps(id string) []questTrackerStep {
	switch id {
	case "coin_in_the_ash":
		return []questTrackerStep{
			{Flag:"found_scratched_coin", Text:"Recover the bronze coin from the cold ash.", Hint:"Search the remains of the cold campfire beneath Olympus."},
			{Flag:"examined_scratched_coin", Text:"Examine the coin's scratched mark.", Hint:"Examine the coin closely; its damage appears deliberate."},
		}
	case "black_thread":
		return []questTrackerStep{
			{Flag:"room_discovered_olympus_foothills", Text:"Follow the black thread into the foothills.", Hint:"Take the mountain road beyond Asterion and look around when you arrive."},
			{Flag:"quest_completed_black_thread", Text:"Resolve the danger guarding the foothills.", Hint:"Inspect the Harpy before deciding whether to fight."},
		}
	case "oracle_whisper":
		return []questTrackerStep{
			{Flag:"room_discovered_delphi_sanctum", Text:"Find the Delphi Sanctum.", Hint:"Follow the road toward Delphi and use 'exits' if the route is unclear."},
			{Flag:"quest_completed_oracle_whisper", Text:"Hear Pythia's warning about the thread.", Hint:"Speak with Pythia about the thread, prophecy, or rebirth."},
		}
	case "river_of_memory":
		return []questTrackerStep{
			{Flag:"quest_completed_oracle_whisper", Text:"Learn what the Oracle knows of the Styx.", Hint:"The Oracle's warning points toward the river."},
			{Flag:"styx_memory_recovered", Text:"Face the choice waiting at the ancient Styx.", Hint:"Travel to the Shores of the Styx and investigate what the crossing asks of you."},
		}
	case "oath_across_the_river":
		return []questTrackerStep{
			{Flag:"oracle_trust", Text:"Accept the Oracle's warning.", Hint:"Listen to the Oracle before deciding what to do with her warning."},
			{Flag:"quest_completed_oath_across_the_river", Text:"Carry the warning to the Temple of the First Dawn.", Hint:"Find the Temple of the First Dawn and enter its grounds."},
		}
	case "unwritten_path":
		return []questTrackerStep{
			{Flag:"oracle_defied", Text:"Choose to defy the Oracle's prophecy.", Hint:"When the Oracle asks for your answer, choose your own path."},
			{Flag:"quest_completed_unwritten_path", Text:"Reach Ancient Athens on your own terms.", Hint:"Follow the road to Ancient Athens without accepting the Oracle's blessing."},
		}
	case "echoes_in_glass":
		return []questTrackerStep{
			{Flag:"styx_memory_recovered", Text:"Carry the Styx memory into another life.", Hint:"The memory must first be recovered at the ancient river."},
			{Flag:"quest_completed_echoes_in_glass", Text:"Find the river beneath modern Athens.", Hint:"In modern Athens, listen for a place where the water sounds strangely familiar."},
		}
	case "museum_echo":
		return []questTrackerStep{
			{Flag:"styx_memory_recovered", Text:"Recover the memory that makes the artifact familiar.", Hint:"The artifact's reflection may make more sense after the ancient Styx."},
			{Flag:"examined_museum_artifact", Text:"Examine the black-glass funerary mask.", Hint:"Find the museum artifact and examine it closely."},
		}
	case "delphi_trials":
		return []questTrackerStep{
			{Flag:"room_discovered_delphi_sanctum", Text:"Discover the Delphi Sanctum.", Hint:"Find the sanctum where the Oracle keeps her broken inscription."},
			{Flag:"examined_delphi_inscription", Text:"Examine the broken inscription.", Hint:"Examine the inscription, then listen for what its silence seems to preserve."},
		}
	case "gate_of_three":
		return []questTrackerStep{
			{Flag:"styx_memory_recovered", Text:"Recover the memory of the Styx.", Hint:"The way to the deepest gate begins with the memory left by the river."},
			{Flag:"cerberus_defeated", Text:"Defeat the guardian at the Gate of Three.", Hint:"Find the gate beneath the river and prepare before challenging its guardian."},
		}
	case "forge_remembers":
		return []questTrackerStep{
			{Flag:"forge_discovered", Text:"Find Theron's Forge.", Hint:"Visit Theron's workshop beyond the foothills."},
			{Flag:"crafted_any_gear", Text:"Craft a piece of equipment.", Hint:"Speak with Theron and use the crafting options available at the forge."},
		}
	case "charons_memory":
		return []questTrackerStep{
			{Flag:"charon_memory_heard", Text:"Hear Charon's account of memory and responsibility.", Hint:"Speak with Charon about memory, the dead, or the river."},
			{Flag:"asphodel_memory_returned", Text:"Bring that meaning to the Unremembered Shade.", Hint:"Find the Unremembered Shade in Asphodel and offer what may help it remember."},
		}
	case "prayer_without_name":
		return []questTrackerStep{
			{Flag:"opened_tomb_coffer", Text:"Open the sealed votive coffer.", Hint:"Search the Manticore Den for a coffer that can be opened."},
			{Flag:"read_charred_prayer", Text:"Read the charred prayer strip.", Hint:"Take the prayer strip from the coffer and read it."},
		}
	case "a_witness_for_the_shade":
		return []questTrackerStep{
			{Flag:"read_harpy_letter", Text:"Read the letter found among the feathers.", Hint:"Inspect the ash the Harpy guards; it may be carrying something other than coins."},
			{Flag:"offered_harpy_letter_to_shade", Text:"Offer the letter to the Unremembered Shade.", Hint:"Carry the letter to Asphodel and offer it to the Shade."},
		}
	case "letter_among_feathers":
		return []questTrackerStep{
			{Flag:"object_taken_harpy_letter", Text:"Recover the feather-folded letter.", Hint:"Search the ash in the foothills after inspecting the Harpy's behavior."},
			{Flag:"read_harpy_letter", Text:"Read the letter.", Hint:"Read the letter from your inventory to discover why it was kept."},
		}
	case "two_cities_one_river":
		return []questTrackerStep{
			{Flag:"listened_village_well", Text:"Listen at Asterion's old well.", Hint:"Return to the village well and listen carefully."},
			{Flag:"heard_modern_styx", Text:"Recognize the same oars beneath modern Athens.", Hint:"In modern Athens, listen near the place where the river's echo returns."},
		}
	case "return_the_coin":
		return []questTrackerStep{
			{Flag:"examined_scratched_coin", Text:"Examine the scratched coin.", Hint:"Examine the coin to understand why it may belong at the well."},
			{Flag:"offered_coin_at_well", Text:"Offer the coin to the clay bowl beside the old well.", Hint:"Return to Asterion's square and offer the coin to the bowl."},
		}
	default:
		return nil
	}
}

func (s *Session) questNextLead(q quest.Quest, steps []questTrackerStep) string {
	for _, step := range steps {
		if !s.Character.HasStoryFlag(step.Flag) { return step.Hint }
	}
	if q.TargetEnemy != "" {
		if s.Character.RoomID == q.TargetRoom {
			return fmt.Sprintf("You are near the danger tied to this quest. Examine %s before choosing how to act.", q.TargetEnemy)
		}
		if room := s.World.GetRoom(q.TargetRoom); room != nil { return fmt.Sprintf("Find your way to %s.", room.Name) }
		return "Follow the clues toward " + strings.ReplaceAll(q.TargetRoom, "_", " ") + "."
	}
	if q.TargetRoom != "" && s.Character.RoomID != q.TargetRoom {
		if room := s.World.GetRoom(q.TargetRoom); room != nil { return fmt.Sprintf("Travel toward %s and pay attention to what the place reveals.", room.Name) }
	}
	if q.RequiredFlag != "" && !s.Character.HasStoryFlag(q.RequiredFlag) { return questUnlockHint(q.RequiredFlag) + "." }
	if q.CompletionFlag != "" && !s.Character.HasStoryFlag(q.CompletionFlag) { return "Examine, read, listen, or offer the relevant object to uncover the next clue." }
	return "Revisit the people and places connected to this story, and examine anything that seems significant."
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
			s.WriteLine("    Reward: %d XP, %d gold", q.RewardXP, q.RewardGold)
		}
		if status == "LOCKED" {
			if q.RequiredFlag != "" && !s.Character.HasStoryFlag(q.RequiredFlag) {
				s.WriteLine("    Unlock: %s", questUnlockHint(q.RequiredFlag))
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
func questUnlockHint(flag string) string {
	switch flag {
	case "opened_tomb_coffer":
		return "open the sealed votive coffer in the Manticore Den and read the prayer strip"
	case "found_scratched_coin":
		return "search the cold campfire ash beneath Olympus"
	case "room_discovered_delphi_sanctum":
		return "find the Delphi Sanctum and read what remains"
	case "styx_memory_recovered":
		return "recover a memory at the ancient Styx"
	case "forge_discovered":
		return "visit Theron's Forge east of the foothills"
	case "charon_memory_heard":
		return "hear Charon explain what a memory can mean to the dead"
	default:
		return "follow the clues that point toward " + strings.ReplaceAll(flag, "_", " ")
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
		s.WriteLine("Use 'current quests' to see active quests, or 'current quest' for your tracked quest.")
	} else {
		s.WriteLine("No new story thread is calling clearly. Explore, revisit old places, or speak with the people you have met.")
	}
	s.WriteLine("")
	s.WriteLine("The journal offers direction, not a leash. You are free to wander.")
}

func (s *Session) tutorial() {
	if s.Character != nil && s.Character.HasStoryFlag("tutorial_available") {
		s.prologueHint()
		s.WriteLine("Your prologue notes are saved. Explore freely, and leave through the one-way gate whenever you feel ready.")
		return
	}
	s.WriteLine("\x1b[1;33mTHE FIRST THREAD — FIELD GUIDE\x1b[0m")
	s.WriteLine("LOOK         'look' describes your current room. It does not list paths; use 'exits' for the detailed route list.")
	s.WriteLine("INSPECT      'look Damon', 'look Harpy', or 'look east' reveals details about a person, creature, or destination.")
	s.WriteLine("OBJECTS      'iscan' lists interactable objects; 'scan' lists people, creatures, and objects.")
	s.WriteLine("COLLECT      'look coin' or 'examine coin' inspects an object; 'get coin from ash' picks it up when takeable.")
	s.WriteLine("MOVE         north/south/east/west (or n/s/e/w) travels along an available path. After moving, type 'look' when you want the room details.")
	s.WriteLine("NAVIGATION   'exits' names each destination; 'map' shows nearby connections; 'worldmap' shows the wider world.")
	s.WriteLine("CONVERSATION 'talk 1 hello' greets NPC 1; 'talk 1 <topic>' asks about a subject. NPC numbers come from 'look'.")
	s.WriteLine("HINTS        'hint' asks the first nearby person for guidance; 'hint 2' selects person 2. Hints are never shown automatically.")
	s.WriteLine("COMBAT       'attack 1' or 'attack harpy' attacks; 'flee' escapes; 'rest' heals when safe. Inspect enemies first with 'look <name>'.")
	s.WriteLine("INVENTORY    'inv' lists carried items; 'eq' lists equipped gear; 'wield <name>' / 'wear <name>' equips items.")
	s.WriteLine("STORY        'quests' tracks objectives; 'journal' gives story context; 'score' shows character progress.")
	s.WriteLine("MERCHANTS    inspect a merchant, use 'shop 1' to view wares, and 'buy 1' to purchase the first item.")
	s.WriteLine("DIVINITY     at level 3 choose one: 'awaken storm', 'awaken tide', 'awaken ember', or 'awaken aegis'.")
	s.WriteLine("REBIRTH      at level 10, use 'rebirth' to cross the Styx into a new era.")
	s.WriteLine("This field guide remains available at any time; the one-time tutorial location itself cannot be revisited after departure.")
}

func (s *Session) storyIntro() {
	if s.Character == nil || s.Character.HasStoryFlag("tutorial_intro_seen") ||
		s.Character.Life != 1 || s.Character.Level > 1 ||
		(s.Character.RoomID != "village_square" && s.Character.RoomID != "village_lane") {
		return
	}
	s.Character.SetStoryFlag("tutorial_intro_seen")
	s.WriteLine("")
	s.WriteLine("\x1b[1;35mTHE THREAD REMEMBERS\x1b[0m")
	s.WriteLine("The first thing you remember is not a face, but the sound of water closing overhead.")
	s.WriteLine("There should be a memory after that: a hand reaching for yours, a prayer spoken too late, perhaps your own voice calling out. Instead there is only the cold, and the stubborn certainty that someone has mistaken the end of your life for the end of your story.")
	s.WriteLine("Three mornings ago, villagers found you beside Asterion's old well. You had no breath and no heartbeat. They had begun preparing a shroud when the dawn reached your face—and you opened your eyes.")
	s.WriteLine("No one agrees on what happened next. Some say you asked for water. Others swear you asked what year it was, then wept when they answered. Damon, who keeps watch on the village lane, says only that a black thread had wound itself around your wrist.")
	s.WriteLine("It pulls faintly toward the mountain. The villagers call that a warning. The priests call it a sign. The old women who leave bread by the well refuse to name it at all.")
	s.WriteLine("You are free to decide what to believe. For now, the village is real beneath your feet, the people have their own troubles, and the road will answer only if you choose to follow it.")
	s.WriteLine("")
}
func (s *Session) prepareTutorial() {
	if s.Character == nil { return }
	if s.Character.HasStoryFlag("tutorial_available") && !s.Character.HasStoryFlag("tutorial_complete") {
		if !s.Character.HasStoryFlag("tutorial_started") {
			s.Character.SetStoryFlag("tutorial_started")
			s.Character.RoomID = "tutorial_shore"
		}
		s.PrologueStep = 0
		for step := 1; step <= 38; step++ {
			if s.Character.HasStoryFlag(fmt.Sprintf("tutorial_lesson_%d", step)) { s.PrologueStep = step }
		}
		s.TutorialStep = 16
		return
	}
	if s.Character.Life == 1 && s.Character.Level <= 1 && s.Character.RoomID == "olympus_gates" &&
		!s.Character.HasStoryFlag("room_discovered_olympus_foothills") && !s.Character.HasStoryFlag("tutorial_complete") {
		s.Character.RoomID = "village_square"
	}
	for step := 16; step >= 0; step-- {
		if step == 16 && s.Character.HasStoryFlag("tutorial_complete") { s.TutorialStep = 16; return }
		if step > 0 && s.Character.HasStoryFlag(fmt.Sprintf("tutorial_step_%d", step)) { s.TutorialStep = step; return }
	}
	// Map characters with progress from the earlier ten-stage tutorial onto the closest equivalent lesson.
	switch {
	case s.Character.HasStoryFlag("tutorial_equipment_complete"): s.TutorialStep = 14
	case s.Character.HasStoryFlag("tutorial_inventory_complete"): s.TutorialStep = 13
	case s.Character.HasStoryFlag("tutorial_combat_complete"): s.TutorialStep = 12
	case s.Character.HasStoryFlag("tutorial_navigation_complete"): s.TutorialStep = 9
	case s.Character.HasStoryFlag("tutorial_hint_complete"): s.TutorialStep = 7
	case s.Character.HasStoryFlag("tutorial_conversation_complete"): s.TutorialStep = 5
	case s.Character.HasStoryFlag("tutorial_inspect_complete"): s.TutorialStep = 3
	case s.Character.HasStoryFlag("tutorial_look_complete"): s.TutorialStep = 2
	case s.Character.HasStoryFlag("tutorial_movement_complete"): s.TutorialStep = 2
	default: s.TutorialStep = 0
	}
}

func (s *Session) prologueHint() {
	lessons := []string{
		"LESSON 1/39 — ORIENTATION: Type 'look'.\nRead the room title, then the description. The title tells you where you are; the prose describes atmosphere, landmarks, and potential points of interest. Room descriptions are not decoration alone: later, clues in the wording may point toward people, exits, dangers, and story discoveries.",
		"LESSON 2/39 — CHARACTER SHEET: Type 'score'.\nRead each value instead of only looking at your level. HP is your health; if it reaches zero, the game restores you but moves you to a recovery location. Mana fuels divine powers later. Experience drives levels, while Life and Era tell you which chapter of your soul you inhabit.",
		"LESSON 3/39 — YOUR PACK: Type 'inv'.\nInventory lists the items your character carries. A carried weapon is not automatically the weapon you are using, and armor in the pack does not protect you until equipped. As you play, inspect names, rarity, damage, armor, and any special relic markers rather than judging an item by its name alone.",
		"LESSON 4/39 — FIRST MOVEMENT: Type 'north' to reach the Wayfarer's Shelter.\nMovement follows exits attached to the current room; it is not free movement across a map. Full directions such as north and east work, and common short forms such as n and e work too. A successful move changes your room, may reveal a new description, and can trigger discovery or story events.",
		"LESSON 5/39 — INSPECT A PERSON: Type 'look Ianthe'.\nThe room view intentionally stays compact. Looking at a named person reveals their fuller description and tells you how to begin interacting. In later rooms, use the name shown by 'look' rather than guessing who might be important.",
		"LESSON 6/39 — GREET THE GUIDE: Type 'talk Ianthe hello'.\nNPCs are addressed by name or by the number shown in a room summary. The topic after the name selects the kind of response you want. The dialogue system recognizes meaningful topic words, so natural questions can work; you do not need to memorize one magic sentence.",
		"LESSON 7/39 — ASK ABOUT MOVEMENT: Type 'talk Ianthe movement'.\nThe world has three related navigation tools: 'look direction' previews a connected room, 'exits' lists exact routes, and 'map' summarizes nearby connections. These tools answer different questions. Previewing a path does not move you, and a map does not grant access to a locked destination.",
		"LESSON 8/39 — LIST EXITS: Type 'exits'.\nRead the directions and destination names. If you are unsure where to go, compare the exit list with the room description and your current objective. A path that is not listed is not available from this room; later, some visible routes may also have level or story restrictions.",
		"LESSON 9/39 — READ THE LOCAL MAP: Type 'map'.\nUse the map to understand the shape of nearby rooms, then use 'exits' when you need the exact destination for a direction. Maps are useful for orientation but do not replace reading descriptions. Automatic maps can be enabled or disabled later with 'automap on' and 'automap off'.",
		"LESSON 10/39 — PREVIEW A DESTINATION: Type 'look east'.\nNotice that the game describes the room to the east while you remain in the shelter. This is a low-risk way to check a route before committing to it. It is especially useful when an exit might lead toward a monster, a merchant, a dead end, or a region you are not yet ready to enter.",
		"LESSON 11/39 — ENTER THE COURTYARD: Type 'east'.\nYou have practiced observing, asking, and planning; now move into the next room. Read the new room description and the threat list. A threat can be visible before you fight it, and seeing a creature is not the same thing as starting combat.",
		"LESSON 12/39 — INSPECT THE TRAINING SHADE: Type 'look Training Shade'.\nBefore attacking any creature, learn its name and read its description. Look for clues about posture, weapons, injuries, or intent. In the main world, some monsters are immediately hostile; the tutorial creatures are deliberately passive so you can decide when to engage.",
		"LESSON 13/39 — FIRST ATTACK: Type 'attack Training Shade'.\nYour attack uses your equipped weapon and character attack values against the enemy's defense. Combat text describes where the blow lands and shows damage in parentheses. The shade will defend itself only after you attack; this is the key difference between a passive creature and an aggressive one.",
		"LESSON 14/39 — VERIFY YOUR PACK: Type 'inv'.\nAfter the fight, confirm that your inventory command works independently of combat. Defeating a tutorial creature records a persistent character flag. Looking around, checking gear, disconnecting, or reconnecting must not bring that slain tutorial creature back.",
		"LESSON 15/39 — EQUIPPED GEAR: Type 'eq'.\nEquipment shows the weapon and armor currently in use, while inventory shows everything you carry. A weapon contributes to your strikes; armor improves defense. If you later find several similar items, use the full inventory names and the equip command's numbered choices to avoid selecting the wrong one.",
		"LESSON 16/39 — SAFE RECOVERY: Type 'rest'.\nRest restores health and mana when no living enemy is actively fighting you. It is a preparation tool, not a way to ignore an active threat. Before a dangerous trip, check your health, rest in safety, and decide whether your current gear is appropriate.",
		"LESSON 17/39 — AUTOMAP ON: Type 'automap on'.\nThe local map will now appear after successful movement. This can be helpful in branching areas, but it adds output to each move. The setting belongs to this session, so you can adjust it to suit the screen and client you are using.",
		"LESSON 18/39 — AUTOMAP OFF: Type 'automap off'.\nYou have seen both styles of navigation output. Keep automatic mapping on when exploring an unfamiliar maze, or turn it off when you prefer concise room descriptions. You can always type 'map' manually whenever you need an overview.",
		"LESSON 19/39 — ENTER THE GROVE: Type 'north'.\nThe next practice encounter is designed to teach intent, not reflex. The creature is visible and wary, but it will not attack you first. Take a moment to notice how the room description communicates that this encounter is under your control.",
		"LESSON 20/39 — INSPECT THE SATYR: Type 'look Satyr'.\nRead the creature's posture and the detail that its staff is lowered. The game expects you to gather information before deciding what to do. In Fatewalker, not every creature must become an enemy; your choice to attack can be the event that turns a peaceful encounter into combat.",
		"LESSON 21/39 — CHOOSE TO ENGAGE: Type 'attack Satyr'.\nYou are choosing to initiate the fight. Once attacked, the satyr becomes aggressive and may retaliate if it survives your strike. In later situations, consider your health, armor, available powers, escape routes, and whether a fight is actually necessary before committing.",
		"LESSON 22/39 — CONFIRM PERSISTENCE: Type 'look'.\nThe room should be quiet after the satyr is slain. Its backend defeat flag belongs to your character, not merely to this connection. That means the tutorial does not force you to repeat the same fight after looking around or reconnecting, and it can safely let you move on.",
		"LESSON 23/39 — RETURN TO THE COURTYARD: Type 'south'.\nRetracing a route is ordinary exploration. Rooms form a network, and you learn that network by combining directions, exits, maps, and descriptions. The tutorial's creatures remain defeated as you revisit their rooms; this is different from repeatable hunting encounters in some parts of the main world.",
		"LESSON 24/39 — ENTER THE STOREHOUSE: Type 'east'.\nThis room brings the practical systems together. You have moved, inspected, spoken, navigated, and fought; now you will review the help system and the tools used to follow a larger story. Read the labels and consider which command you would use if you forgot a mechanic later.",
		"LESSON 25/39 — INTERACTABLE OBJECTS: Type 'iscan', then 'open practice chest' and 'look in practice chest'.\nThe room contains real objects, not only descriptive text. 'iscan' lists things you can handle, containers, and scenery. A closed container hides its contents until you open it; 'look in' lets you inspect what is inside. Type 'inv' when you are ready to continue.",
		"LESSON 26/39 — PICK UP AND EXAMINE: Type 'get practice bandage', then 'examine practice bandage', then 'eq'.\n'get' moves a takeable room object into your inventory; 'examine' gives a closer description. The bandage is a real consumable, and its note explains how to use it. Inventory shows what you carry; equipment shows what you have actively equipped. To change gear later, use 'wield <item>', 'wear <item>', or 'equip <item>'.",
		"LESSON 27/39 — READ COMBAT HELP: Type 'help combat'.\nHelp files explain rules, command syntax, examples, and common mistakes. Use 'help list' to discover topics. Important guides include movement, combat, guard, consumables, crafting, progression, quests, rebirth, and soul legacy. You do not need to memorize every mechanic; learn how to find the correct guide when you need it.",
		"LESSON 28/39 — WIDER NAVIGATION: Type 'worldmap'.\nThe local map focuses on your immediate neighborhood; the world map helps you think about the larger setting. Future regions may be locked until you reach a required life, level, or story condition. A visible destination is not necessarily one you can enter right now.",
		"LESSON 29/39 — RETURN TO THE COURTYARD: Type 'west'.\nUse the exit network to return. In a larger area, you may need to check 'exits' or 'map' again rather than assuming every room connects in the same pattern. Exploration is deliberate: paths, room descriptions, and quest clues work together.",
		"LESSON 30/39 — RETURN TO THE SHELTER: Type 'west'.\nIanthe remains available for questions. The room summary gives NPC names and numbers; you can use either when talking or asking for a hint. Later, merchants and story characters may offer different interactions, so inspect them and read their prompts.",
		"LESSON 31/39 — ASK ABOUT COMBAT: Type 'talk Ianthe combat'.\nNPCs can provide lore, directions, hints, and quest context. Ask about the subject you need rather than repeating hello forever. Dialogue is not a universal AI: it matches known topics and aliases, so try a related keyword if the character does not recognize your first question.",
		"LESSON 32/39 — QUEST TRACKING: Type 'quests'.\nQuests show active, completed, and locked objectives. Some advance when you reach a room; others require a specific enemy, discovery, or choice. A locked quest is not broken simply because it cannot progress yet—read its unlock hint and explore the story conditions.",
		"LESSON 33/39 — STORY CONTEXT: Type 'journal'.\nThe journal summarizes the larger thread of your story, while quests provide more specific objectives. Neither is meant to remove your freedom. You can follow the suggested path, wander to find secrets, talk to other characters, or return to a mystery when you are stronger.",
		"LESSON 34/39 — CHECK PROGRESS: Type 'score'.\nRead HP, mana, XP, attack, defense, armor, gold, life, and era. Level-ups strengthen your character. At level 3, choose one domain with 'awaken storm', 'awaken tide', 'awaken ember', or 'awaken aegis'; powers then unlock at their listed levels. At level 10, rebirth opens. It resets level and XP, but the soul carries its history and relics.",
		"LESSON 35/39 — SURVIVAL AND COMBAT CHOICES: Type 'east'.\nIn a real fight, 'attack' deals weapon damage, 'cast <power>' spends mana for a divine attack, 'use <item>' consumes a healing or mana supply, 'guard' reduces the next incoming hit, and 'flee' abandons the encounter. These choices trade damage, resources, and safety; read 'help combat' and 'help guard' before a difficult battle.",
		"LESSON 36/39 — ECONOMY AND CRAFTING: Type 'east'.\nVictories and discoveries can earn XP and gold. Merchants sell equipment and supplies; use 'shop' then 'buy <item>'. Some enemies leave crafting materials marked [MATERIAL]. Carry the named ingredient and enough gold to Theron's Forge, then type 'craft' to see recipes. Crafted gear goes into your inventory and must be equipped. Read 'help economy' and 'help crafting' for details.",
		"LESSON 37/39 — APPROACH THE GATE: Type 'north'.\nThe marble gate is open whenever you feel ready to leave. There is no penalty for skipping a lesson or taking a detour; the notes and help files will remain available after you leave.",
		"LESSON 38/39 — PREVIEW THE REAL WORLD: Type 'look north'.\nThe destination preview describes Asterion Village without moving you there. Read the description and imagine the first choices you might make: talk to the village guide, investigate the road, explore the forge, or follow the main story. The gate is the boundary between safe practice and the open game.",
		"LESSON 39/39 — DEPART: Type 'north'.\nThis is the threshold between a safe shelter and a living world. When you leave, the game records the prologue as complete and places you in Asterion Village. The shelter will close behind you, but the wider world will not demand that you follow one prescribed route.",
	}
	if s.PrologueStep >= len(lessons) { s.WriteLine("The prologue is complete. Continue into Fatewalker proper."); return }
	s.WriteLine("\x1b[1;36mTHE FIRST THREAD — WAYFARER FIELD NOTES\x1b[0m")
	parts := strings.SplitN(lessons[s.PrologueStep], "\n", 2)
	heading := strings.TrimSpace(parts[0])
	if divider := strings.Index(heading, " — "); divider >= 0 {
		heading = strings.TrimSpace(heading[divider+len(" — "):])
	}
	s.WriteLine("\x1b[1;33m%s\x1b[0m", heading)
	if len(parts) > 1 { s.WriteLine("%s", parts[1]) }
	s.WriteLine("This is a suggestion, not a test. Explore, talk, inspect, and experiment in any order; the gate is available when you feel ready.")
	s.WriteLine("Progress is saved automatically. Type 'tutorial' for these notes, or 'help <topic>' for practical details.")
}

func (s *Session) advancePrologueForCommand(p []string) {
	if len(p)==0 || s.Character==nil || !s.Character.HasStoryFlag("tutorial_available") { return }
	cmd, args, room := p[0], strings.Join(p[1:], " "), s.Character.RoomID
	has := func(words ...string) bool { for _, w := range words { if strings.Contains(strings.ToLower(args), w) { return true } }; return false }
	next := -1
	switch s.PrologueStep {
	case 0: if cmd=="look" && len(p)==1 && room=="tutorial_shore" { next=1 }
	case 1: if (cmd=="score"||cmd=="stats") && room=="tutorial_shore" { next=2 }
	case 2: if (cmd=="inv"||cmd=="inventory"||cmd=="i") && room=="tutorial_shore" { next=3 }
	case 3: if (cmd=="north"||cmd=="n") && room=="tutorial_camp" { next=4 }
	case 4: if cmd=="look" && has("ianthe","guide") && room=="tutorial_camp" { next=5 }
	case 5: if (cmd=="talk"||cmd=="say") && has("hello","greeting") && room=="tutorial_camp" { next=6 }
	case 6: if (cmd=="talk"||cmd=="say") && has("movement","move","directions","exits") && room=="tutorial_camp" { next=7 }
	case 7: if cmd=="exits" && room=="tutorial_camp" { next=8 }
	case 8: if cmd=="map" && room=="tutorial_camp" { next=9 }
	case 9: if cmd=="look" && has("east"," e") && room=="tutorial_camp" { next=10 }
	case 10: if (cmd=="east"||cmd=="e") && room=="tutorial_yard" { next=11 }
	case 11: if cmd=="look" && has("shade","training") && room=="tutorial_yard" { next=12 }
	case 12: if (cmd=="attack"||cmd=="kill"||cmd=="hit") && room=="tutorial_yard" && s.Character.HasStoryFlag("tutorial_enemy_slain_tutorial_yard") { next=13 }
	case 13: if (cmd=="inv"||cmd=="inventory"||cmd=="i") && room=="tutorial_yard" { next=14 }
	case 14: if (cmd=="eq"||cmd=="equipment"||cmd=="gear") && room=="tutorial_yard" { next=15 }
	case 15: if cmd=="rest" && room=="tutorial_yard" { next=16 }
	case 16: if cmd=="automap" && has("on") && room=="tutorial_yard" { next=17 }
	case 17: if cmd=="automap" && has("off") && room=="tutorial_yard" { next=18 }
	case 18: if (cmd=="north"||cmd=="n") && room=="tutorial_grove" { next=19 }
	case 19: if cmd=="look" && has("satyr") && room=="tutorial_grove" { next=20 }
	case 20: if (cmd=="attack"||cmd=="kill"||cmd=="hit") && room=="tutorial_grove" && s.Character.HasStoryFlag("tutorial_enemy_slain_tutorial_grove") { next=21 }
	case 21: if cmd=="look" && len(p)==1 && room=="tutorial_grove" { next=22 }
	case 22: if (cmd=="south"||cmd=="s") && room=="tutorial_yard" { next=23 }
	case 23: if (cmd=="east"||cmd=="e") && room=="tutorial_storehouse" { next=24 }
	case 24: if (cmd=="inv"||cmd=="inventory"||cmd=="i") && room=="tutorial_storehouse" { next=25 }
	case 25: if (cmd=="eq"||cmd=="equipment"||cmd=="gear") && room=="tutorial_storehouse" { next=26 }
	case 26: if cmd=="help" && has("combat") && room=="tutorial_storehouse" { next=27 }
	case 27: if (cmd=="worldmap"||cmd=="world-map") && room=="tutorial_storehouse" { next=28 }
	case 28: if (cmd=="west"||cmd=="w") && room=="tutorial_yard" { next=29 }
	case 29: if (cmd=="west"||cmd=="w") && room=="tutorial_camp" { next=30 }
	case 30: if (cmd=="talk"||cmd=="say") && has("combat","fight","enemy") && room=="tutorial_camp" { next=31 }
	case 31: if cmd=="quests" && room=="tutorial_camp" { next=32 }
	case 32: if (cmd=="journal"||cmd=="quest"||cmd=="story") && room=="tutorial_camp" { next=33 }
	case 33: if (cmd=="score"||cmd=="stats") && room=="tutorial_camp" { next=34 }
	case 34: if (cmd=="east"||cmd=="e") && room=="tutorial_yard" { next=35 }
	case 35: if (cmd=="east"||cmd=="e") && room=="tutorial_storehouse" { next=36 }
	case 36: if (cmd=="north"||cmd=="n") && room=="tutorial_gate" { next=37 }
	case 37: if cmd=="look" && has("north"," n") && room=="tutorial_gate" { next=38 }
	case 38:
		if (cmd=="north"||cmd=="n") && room=="village_square" {
			s.Character.SetStoryFlag("tutorial_complete")
			s.Character.SetStoryFlag("tutorial_finished")
			s.Character.StoryFlags["tutorial_available"] = false
			s.PrologueStep=39
			s.WriteLine("\x1b[1;32mTHE FIRST THREAD — COMPLETE\x1b[0m")
			s.WriteLine("The marble gate seals behind you. Your training is remembered, but the tutorial itself is no longer reachable.")
			s.WriteLine("You have entered Fatewalker proper. Use 'help list', 'journal', 'quests', and 'look' as you explore.")
			return
		}
	}
	if next>=0 { s.PrologueStep=next; s.Character.SetStoryFlag(fmt.Sprintf("tutorial_lesson_%d",next)); s.prologueHint() }
}

func (s *Session) tutorialHint() {
	if s.Character != nil && s.Character.HasStoryFlag("tutorial_available") { s.prologueHint(); return }
	switch s.TutorialStep {
	case 0: s.WriteLine("\x1b[1;36mTUTORIAL 1/16 — LOOK:\x1b[0m Type 'look'.")
	case 1: s.WriteLine("\x1b[1;36mTUTORIAL 2/16 — MOVE:\x1b[0m Now go north to the village lane.")
	case 2: s.WriteLine("\x1b[1;36mTUTORIAL 3/16 — INSPECT DAMON:\x1b[0m Type 'look Damon' to learn more about the guide.")
	case 3: s.WriteLine("\x1b[1;36mTUTORIAL 4/16 — GREET DAMON:\x1b[0m Type 'talk Damon hello' (or use the number selector: 'talk 1 hello').")
	case 4: s.WriteLine("\x1b[1;36mTUTORIAL 5/16 — ASK ABOUT THE VILLAGE:\x1b[0m Type 'talk Damon village' (or 'talk 1 village').")
	case 5: s.WriteLine("\x1b[1;36mTUTORIAL 6/16 — REQUEST GUIDANCE:\x1b[0m Type 'hint' to ask Damon for a hint.")
	case 6: s.WriteLine("\x1b[1;36mTUTORIAL 7/16 — ASK ABOUT THE ROAD:\x1b[0m Type 'talk Damon road' (or 'talk 1 road') to learn about the mountain threat.")
	case 7: s.WriteLine("\x1b[1;36mTUTORIAL 8/16 — READ EXITS:\x1b[0m Type 'exits' to see the detailed list of paths.")
	case 8: s.WriteLine("\x1b[1;36mTUTORIAL 9/16 — VIEW THE MAP:\x1b[0m Type 'map' to see nearby room connections.")
	case 9: s.WriteLine("\x1b[1;36mTUTORIAL 10/16 — FOLLOW THE ROAD:\x1b[0m Now go north to the foothills.")
	case 10: s.WriteLine("\x1b[1;36mTUTORIAL 11/16 — INSPECT THE HARPY:\x1b[0m Type 'look Harpy' before attacking.")
	case 11: s.WriteLine("\x1b[1;36mTUTORIAL 12/16 — FIGHT:\x1b[0m Type 'attack Harpy' (or use the number selector: 'attack 1').")
	case 12: s.WriteLine("\x1b[1;36mTUTORIAL 13/16 — INVENTORY:\x1b[0m Type 'inv' to review what you carry.")
	case 13: s.WriteLine("\x1b[1;36mTUTORIAL 14/16 — EQUIPMENT:\x1b[0m Type 'eq' to review what you have equipped.")
	case 14: s.WriteLine("\x1b[1;36mTUTORIAL 15/16 — QUESTS:\x1b[0m Type 'quests' to review your current objectives.")
	case 15: s.WriteLine("\x1b[1;36mTUTORIAL 16/16 — JOURNAL:\x1b[0m Type 'journal' to read the story context.")
	case 16:
		s.WriteLine("\x1b[1;32mTUTORIAL COMPLETE:\x1b[0m You have practiced room inspection, movement, conversation, hints, navigation, combat, inventory, equipment, and story tracking.")
		s.WriteLine("Explore freely. Use 'tutorial' or 'help <topic>' whenever you need a refresher.")
	}
}
func (s *Session) advanceTutorial(step int) {
	if step <= s.TutorialStep { return }
	s.TutorialStep = step
	s.Character.SetStoryFlag(fmt.Sprintf("tutorial_step_%d", step))
	switch step {
	case 2: s.Character.SetStoryFlag("tutorial_movement_complete")
	case 3: s.Character.SetStoryFlag("tutorial_inspect_complete")
	case 5: s.Character.SetStoryFlag("tutorial_conversation_complete")
	case 7: s.Character.SetStoryFlag("tutorial_hint_complete")
	case 9: s.Character.SetStoryFlag("tutorial_navigation_complete")
	case 12: s.Character.SetStoryFlag("tutorial_combat_complete")
	case 13: s.Character.SetStoryFlag("tutorial_inventory_complete")
	case 14: s.Character.SetStoryFlag("tutorial_equipment_complete")
	case 16: s.Character.SetStoryFlag("tutorial_complete")
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
			s.WriteLine("Damon's hint: begin a conversation with 'talk Damon hello'. You can also use the number selector: 'talk 1 hello'.")
		case 5, 6:
			s.Character.SetStoryFlag("tutorial_hint_used")
			s.WriteLine("Damon's hint: ask about the road with 'talk Damon road', or use the number selector: 'talk 1 road'.")
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
	if r == nil { s.WriteLine("\x1b[1;31mYou are nowhere.\x1b[0m The world has lost track of you."); return }
	s.WriteLine("")
	s.WriteLine("\x1b[1;36m╭─ %s ─╮\x1b[0m", r.Name)
	s.WriteLine("%s", r.Description)
	if visible := visibleRoomObjects(r, s.Character); len(visible) > 0 {
		s.WriteLine("")
		s.WriteLine("\x1b[1;37mOBJECTS & DETAILS\x1b[0m")
		for _, object := range visible {
			label := ""
			if object.Container { label = " \x1b[90m[container]\x1b[0m" } else if object.Takeable { label = " \x1b[32m[takeable]\x1b[0m" } else { label = " \x1b[90m[scenery]\x1b[0m" }
			if object.ContainerID != "" {
				if parent := findRoomObjectByID(r, object.ContainerID); parent != nil {
					s.WriteLine("  %s%s \x1b[90m(in %s)\x1b[0m", object.Name, label, parent.Name)
					continue
				}
			}
			s.WriteLine("  %s%s", object.Name, label)
		}
	}
	others := s.onlineSnapshotList()
	shownOthers := 0
	for _, player := range others {
		if strings.EqualFold(player.Name, s.Character.Name) || player.RoomID != s.Character.RoomID { continue }
		if shownOthers == 0 { s.WriteLine(""); s.WriteLine("\x1b[1;37mOTHER TRAVELERS\x1b[0m") }
		shownOthers++
		s.WriteLine("  \x1b[1;36m%s\x1b[0m — Level %d, Life %d", player.Name, player.Level, player.Life)
	}
	npcs := s.World.NPCs[s.Character.RoomID]
	if len(npcs) > 0 {
		s.WriteLine("")
		s.WriteLine("\x1b[1;37mPEOPLE & INTERACTABLES\x1b[0m")
		for i, n := range npcs {
			label := ""
			if n.HasShop() { label = " \x1b[33m[MERCHANT]\x1b[0m" }
			s.WriteLine("  \x1b[32m[NPC %d] %s\x1b[0m%s", i+1, n.Name, label)
		}
	}
	if s.Enemy == nil && s.EncounterClearedRoom != s.Character.RoomID { s.spawnEnemy() }
	if s.Enemy != nil && s.Enemy.HP > 0 {
		s.WriteLine("")
		s.WriteLine("\x1b[1;33mTHREATS\x1b[0m")
		s.WriteLine("  \x1b[1;33m[ENEMY 1] %s\x1b[0m — Health %d/%d", s.Enemy.Name, s.Enemy.HP, s.Enemy.MaxHP)
	} else if s.EncounterClearedRoom == s.Character.RoomID {
		s.WriteLine("")
		s.WriteLine("\x1b[1;32mTHREATS\x1b[0m")
		s.WriteLine("  The area is quiet. The defeated creature has not returned.")
	}
}

func (s *Session) lookTarget(args []string) {
	if s.Character == nil || len(args) == 0 { s.look(); return }
	target := strings.ToLower(strings.TrimSpace(strings.Join(args, " ")))
	target = strings.TrimPrefix(target, "at ")
	target = strings.TrimPrefix(target, "in ")
	target = stripLeadingArticle(target)
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil { s.WriteLine("You are nowhere."); return }
	directions := map[string]string{"n":"north", "s":"south", "e":"east", "w":"west", "u":"up", "d":"down", "north":"north", "south":"south", "east":"east", "west":"west", "up":"up", "down":"down", "in":"in", "out":"out"}
	if direction, ok := directions[target]; ok {
		if id, exists := r.Exits[direction]; exists {
			if destination := s.World.GetRoom(id); destination != nil {
				s.WriteLine("\x1b[1;36m%s\x1b[0m", destination.Name)
				s.WriteLine("%s", destination.Description)
			} else { s.WriteLine("The path leads toward %s.", id) }
		} else { s.WriteLine("There is no visible path %s from here.", direction) }
		return
	}
	if object := findRoomObject(r, target, "", s.Character); object != nil {
		s.describeObject(object, r, s.Character)
		return
	}
	for _, carried := range s.Character.Inventory {
		if strings.Contains(strings.ToLower(carried.Name), target) || strings.EqualFold(carried.Name, target) {
			s.WriteLine("\x1b[1;36m%s\x1b[0m", carried.Name)
			if carried.Description != "" { s.WriteLine("%s", carried.Description) } else { s.WriteLine("A %s you are carrying.", carried.Kind) }
			s.WriteLine("Quality: %s", carried.TierName())
			if carried.Damage > 0 { s.WriteLine("Weapon damage: %d", carried.Damage) }
			if carried.Armor > 0 { s.WriteLine("Armor: %d", carried.Armor) }
			if carried.Relic { s.WriteLine("This is a soul relic and may endure beyond rebirth.") }
			return
		}
	}
	for _, container := range r.Objects {
		if container == nil || !container.Container || !container.Open || !container.VisibleIn(r.Objects) { continue }
		for _, stored := range s.Character.ContainerStorage[containerStorageKey(r.ID, container.ID)] {
			if strings.Contains(strings.ToLower(stored.Name), target) {
				s.WriteLine("\x1b[1;36m%s\x1b[0m", stored.Name)
				if stored.Description != "" { s.WriteLine("%s", stored.Description) }
				s.WriteLine("It is stored inside %s.", container.Name)
				s.WriteLine("Use 'get %s from %s' to retrieve it.", strings.ToLower(stored.Name), strings.ToLower(container.Name))
				return
			}
		}
	}
	for i, n := range s.World.NPCs[s.Character.RoomID] {
		name := strings.ToLower(n.Name)
		if target == strconv.Itoa(i+1) || strings.Contains(name, target) || strings.Contains(target, name) {
			s.WriteLine("\x1b[1;36m%s\x1b[0m", n.Name)
			if n.Description != "" { s.WriteLine("%s", n.Description) }
			if s.TutorialStep == 2 && n.ID == "village_guide" { s.Character.SetStoryFlag("tutorial_inspected_npc") }
			if n.Faction != "" { s.WriteLine("Affiliation: %s", strings.ReplaceAll(n.Faction, "_", " ")) }
			if n.HasShop() { s.WriteLine("This character trades in equipment. Use 'shop %d' to inspect their wares.", i+1) }
			s.WriteLine("Speak with 'talk %d <topic>' or ask for guidance with 'hint %d'.", i+1, i+1)
			return
		}
	}
	if s.Enemy == nil && s.EncounterClearedRoom != s.Character.RoomID { s.spawnEnemy() }
	if s.Enemy != nil && s.Enemy.HP > 0 {
		name := strings.ToLower(s.Enemy.Name)
		if strings.Contains(name, target) || strings.Contains(target, name) || target == "enemy" || target == "threat" || target == "mob" {
			s.WriteLine("\x1b[1;33m%s\x1b[0m", s.Enemy.Name)
			if s.Enemy.Description != "" { s.WriteLine("%s", s.Enemy.Description) }
			if (s.TutorialStep == 10 || s.TutorialStep == 11) && s.Enemy.Name == "Harpy" { s.Character.SetStoryFlag("tutorial_enemy_inspected") }
			s.WriteLine("Health: %d / %d", s.Enemy.HP, s.Enemy.MaxHP)
			s.WriteLine("It is hostile. Use 'attack 1' or 'attack %s' to engage.", strings.ToLower(s.Enemy.Name))
			return
		}
	}
	s.WriteLine("You see no person, creature, or path matching '%s'. Try 'look' for the room summary.", target)
}

func (s *Session) listen(args []string) {
	if len(args) == 0 {
		s.WriteLine("You pause and listen. Name what you want to hear, such as 'listen well' or 'listen to the water'.")
		return
	}
	target := strings.ToLower(strings.Join(args, " "))
	target = strings.TrimPrefix(target, "to ")
	switch s.Character.RoomID {
	case "village_square":
		if strings.Contains(target, "well") || strings.Contains(target, "water") || strings.Contains(target, "bowl") {
			if !s.Character.HasStoryFlag("listened_village_well") {
				s.Character.SetStoryFlag("listened_village_well")
				s.WriteLine("You lean over the old well. Beneath the soft movement of water, you hear a second sound: oars dipping into a river, followed by a passenger whispering a name you almost recognize. The sound comes from below the stone, not from the water itself.")
				s.WriteLine("When you straighten, the village sounds return — a cart wheel, a dog, someone calling for bread. Yet the rhythm of the oars stays with you.")
				s.updateQuests()
				s.persist()
			} else {
				s.WriteLine("The well gives you only its quiet rhythm now. Once heard, the distant oars are difficult to mistake for ordinary water.")
			}
			return
		}
	case "modern_styx":
		if strings.Contains(target, "water") || strings.Contains(target, "river") || strings.Contains(target, "tunnel") || strings.Contains(target, "track") || strings.Contains(target, "platform") {
			if !s.Character.HasStoryFlag("listened_village_well") {
				s.WriteLine("The dark water murmurs beneath the station, but the sound has no pattern you can follow. It feels as if you have arrived in the middle of a conversation.")
				s.WriteLine("Perhaps you have heard its beginning somewhere else.")
				return
			}
			if !s.Character.HasStoryFlag("heard_modern_styx") {
				s.Character.SetStoryFlag("heard_modern_styx")
				s.Character.Memories = append(s.Character.Memories, "Beneath modern Athens, I heard the same oars that sounded under Asterion's old well. The river connects places that should never touch.")
				s.WriteLine("You close your eyes. The station's electrical hum falls away, and the oars return — the same measured dip, the same pause, the same whisper from the well. The two places are not imitating one another. They are listening to the same river.")
				s.WriteLine("A train passes overhead. For one impossible second, every window reflects a black shoreline under a starless sky.")
				s.updateQuests()
				s.persist()
			} else {
				s.WriteLine("The river's rhythm is familiar now. The city roars above it, but the oars continue at their patient, impossible pace.")
			}
			return
		}
	}
	s.WriteLine("You listen carefully, but nothing nearby answers in a way you can place. Try naming a sound or feature in the room.")
}

func (s *Session) examineTarget(args []string) {
	if len(args) == 0 { s.look(); return }
	s.lookTarget(args)
	query := strings.ToLower(strings.Join(args, " "))
	if s.Character.RoomID == "delphi_sanctum" && (strings.Contains(query, "inscription") || strings.Contains(query, "writing") || strings.Contains(query, "stone")) {
		s.Character.SetStoryFlag("examined_delphi_inscription")
		s.updateQuests()
	}
	if s.Character.RoomID == "modern_museum" && (strings.Contains(query, "artifact") || strings.Contains(query, "mask") || strings.Contains(query, "glass")) {
		s.Character.SetStoryFlag("examined_museum_artifact")
		s.updateQuests()
	}
	if strings.Contains(query, "coin") || strings.Contains(query, "bronze") || strings.Contains(query, "scratched") {
		for _, carried := range s.Character.Inventory {
			if strings.Contains(strings.ToLower(carried.Name), "scratched bronze coin") {
				s.Character.SetStoryFlag("examined_scratched_coin")
				s.WriteLine("The scratched mark is not random damage. It resembles a thread tied in a knot, the same sign that circles your wrist.")
				s.updateQuests()
				break
			}
		}
	}
}

func containerStorageKey(roomID, containerID string) string {
	return roomID + "/" + containerID
}

func visibleRoomObjects(r *world.Room, character *character.Character) []*world.Object {
	if r == nil { return nil }
	var found []*world.Object
	for _, object := range r.Objects {
		if object != nil && object.VisibleIn(r.Objects) && (character == nil || !character.HasStoryFlag("object_taken_"+object.ID)) { found = append(found, object) }
	}
	return found
}

func findRoomObjectByID(r *world.Room, id string) *world.Object {
	if r == nil { return nil }
	for _, object := range r.Objects { if object != nil && object.ID == id { return object } }
	return nil
}

func stripLeadingArticle(value string) string {
	for _, article := range []string{"the ", "a ", "an "} {
		value = strings.TrimPrefix(strings.TrimSpace(value), article)
	}
	return strings.TrimSpace(value)
}

func findRoomObject(r *world.Room, query, source string, character *character.Character) *world.Object {
	if r == nil { return nil }
	query = stripLeadingArticle(strings.TrimSpace(strings.ToLower(query)))
	source = stripLeadingArticle(strings.TrimSpace(strings.ToLower(source)))
	for _, object := range r.Objects {
		if object == nil || !object.VisibleIn(r.Objects) || (character != nil && character.HasStoryFlag("object_taken_"+object.ID)) || !object.Matches(query) { continue }
		if source != "" {
			parent := findRoomObjectByID(r, object.ContainerID)
			if parent == nil || !parent.Matches(source) { continue }
		}
		return object
	}
	return nil
}

func (s *Session) describeObject(object *world.Object, r *world.Room, character *character.Character) {
	if object == nil { return }
	s.WriteLine("\x1b[1;36m%s\x1b[0m", object.Name)
	if object.Description != "" { s.WriteLine("%s", object.Description) }
	if object.ContainerID != "" {
		if parent := findRoomObjectByID(r, object.ContainerID); parent != nil {
			s.WriteLine("It rests in %s.", parent.Name)
		}
	}
	if object.Container {
		if !object.Open {
			s.WriteLine("It is closed. You will need to open it before you can see what lies inside.")
		} else {
			var contents []string
			for _, child := range r.Objects {
				if child != nil && child.ContainerID == object.ID && child.Visible && (character == nil || !character.HasStoryFlag("object_taken_"+child.ID)) { contents = append(contents, child.Name) }
			}
			known := map[string]bool{}
			for _, name := range contents { known[strings.ToLower(name)] = true }
			if character != nil {
				for _, stored := range character.ContainerStorage[containerStorageKey(r.ID, object.ID)] {
					if !known[strings.ToLower(stored.Name)] {
						contents = append(contents, stored.Name)
						known[strings.ToLower(stored.Name)] = true
					}
				}
			}
			if len(contents) > 0 { s.WriteLine("Within it you can make out: %s.", strings.Join(contents, ", ")) } else { s.WriteLine("It appears to be empty.") }
		}
	}
	if object.Takeable { s.WriteLine("You could pick it up with 'get %s'.", strings.ToLower(object.Name)) } else { s.WriteLine("It is part of the place and cannot be taken.") }
}

func (s *Session) openContainer(args []string) {
	if len(args) == 0 { s.WriteLine("Open what? Use 'iscan' to find containers."); return }
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil { s.WriteLine("There is nowhere here to open anything."); return }
	object := findRoomObject(r, strings.Join(args, " "), "", s.Character)
	if object == nil { s.WriteLine("You see no such object here."); return }
	if !object.Container { s.WriteLine("%s is not a container.", object.Name); return }
	if object.Open { s.WriteLine("%s is already open.", object.Name); return }
	object.Open = true
	s.WriteLine("You work the stubborn clasp loose. %s opens with a dry wooden sigh.", object.Name)
	if object.ID == "tomb_coffer" {
		s.Character.SetStoryFlag("opened_tomb_coffer")
		s.WriteLine("Inside lies a narrow strip of charred parchment. The coffer was not locked against thieves; it was waiting for someone to choose to look.")
		s.updateQuests()
	}
	s.persist()
}

func (s *Session) closeContainer(args []string) {
	if len(args) == 0 { s.WriteLine("Close what?"); return }
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil { s.WriteLine("There is nowhere here to close anything."); return }
	object := findRoomObject(r, strings.Join(args, " "), "", s.Character)
	if object == nil { s.WriteLine("You see no such object here."); return }
	if !object.Container { s.WriteLine("%s is not a container.", object.Name); return }
	if !object.Open { s.WriteLine("%s is already closed.", object.Name); return }
	object.Open = false
	s.WriteLine("You close %s.", object.Name)
	s.persist()
}

func (s *Session) searchObject(args []string) {
	if len(args) == 0 { s.WriteLine("Search what? Name a container, fixture, or object."); return }
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil { s.WriteLine("There is nowhere here to search."); return }
	query := strings.Join(args, " ")
	object := findRoomObject(r, query, "", s.Character)
	if object == nil {
		s.WriteLine("You find nothing matching '%s'. Try 'iscan' to list visible objects.", query)
		return
	}
	s.WriteLine("You search %s carefully.", object.Name)
	if object.Container {
		if !object.Open {
			s.WriteLine("The container is closed. Open it before searching inside.")
			return
		}
		found := 0
		for _, child := range r.Objects {
			if child == nil || child.ContainerID != object.ID || !child.Visible || s.Character.HasStoryFlag("object_taken_"+child.ID) { continue }
			found++
			s.WriteLine("  You find: %s.", child.Name)
		}
		for _, stored := range s.Character.ContainerStorage[containerStorageKey(r.ID, object.ID)] {
			found++
			s.WriteLine("  You find: %s.", stored.Name)
		}
		if found == 0 { s.WriteLine("There is nothing else inside.") }
		return
	}
	s.describeObject(object, r, s.Character)
}

func (s *Session) readObject(args []string) {
	if len(args) == 0 { s.WriteLine("Read what? Try 'read prayer strip'."); return }
	query := strings.ToLower(strings.Join(args, " "))
	found, isPrayerStrip, isHarpyLetter := false, false, false
	for _, carried := range s.Character.Inventory {
		name := strings.ToLower(carried.Name)
		if strings.Contains(name, query) || strings.Contains(query, name) {
			found = true
			if strings.Contains(name, "charred prayer strip") { isPrayerStrip = true }
			if strings.Contains(name, "feather-folded letter") { isHarpyLetter = true }
		}
	}
	if !found {
		r := s.World.GetRoom(s.Character.RoomID)
		if object := findRoomObject(r, query, "", s.Character); object != nil {
			found = true
			isPrayerStrip = object.ID == "charred_prayer_strip"
			isHarpyLetter = object.ID == "harpy_letter"
		}
	}
	if !found { s.WriteLine("You cannot find anything readable matching '%s'.", query); return }
	if isHarpyLetter {
		if !s.Character.HasStoryFlag("object_taken_harpy_letter") {
			s.WriteLine("The letter is caught beneath a feather. Pick it up first if you want to unfold it safely.")
			return
		}
		if !s.Character.HasStoryFlag("read_harpy_letter") {
			s.Character.SetStoryFlag("read_harpy_letter")
			s.Character.Memories = append(s.Character.Memories, "I read the feather-folded letter: its writer begged someone to tell a sister that staying behind was a choice, not abandonment.")
			s.WriteLine("The letter begins without a greeting. The writer says they stayed at the sealed mountain shrine so a younger sister could escape before the creatures came down the road.")
			s.WriteLine("The final lines are smudged, but one sentence survives: 'If they tell you I was taken, do not believe them. I stayed because someone had to hold the door.' There is no name, only the three-line mark.")
			s.updateQuests()
			s.persist()
		} else {
			s.WriteLine("You unfold the letter again. Its last surviving sentence still refuses to sound like a farewell.")
		}
		return
	}
	if isPrayerStrip {
		if !s.Character.HasStoryFlag("opened_tomb_coffer") {
			s.WriteLine("The words remain hidden. You must first open the coffer in the Manticore Den.")
			return
		}
		if !s.Character.HasStoryFlag("read_charred_prayer") {
			s.Character.SetStoryFlag("read_charred_prayer")
			s.Character.Memories = append(s.Character.Memories, "I read the charred prayer strip: someone asked not to be remembered, only for another to remember why they stayed.")
			s.WriteLine("You read the surviving words slowly. They do not name a god, a hero, or a king. They describe someone who stayed beside a frightened person when every sensible instinct urged them to flee.")
			s.WriteLine("The three-line mark beneath the words is not a signature. It is a witness mark — proof that someone else was there.")
			s.updateQuests()
			s.persist()
		} else {
			s.WriteLine("You read the prayer again. The words have not changed, but the choice they describe feels heavier now.")
		}
		return
	}
	s.examineTarget(args)
}

func (s *Session) getObject(args []string) {
	if len(args) == 1 && strings.EqualFold(args[0], "all") {
		r := s.World.GetRoom(s.Character.RoomID)
		if r == nil { s.WriteLine("There is nowhere here to search."); return }
		objects := visibleRoomObjects(r, s.Character)
		picked := 0
		for _, object := range objects {
			if object.Takeable {
				s.getObject([]string{object.Name})
				picked++
			}
		}
		if picked == 0 { s.WriteLine("There is nothing takeable here.") }
		return
	}
	if len(args) == 0 { s.WriteLine("Get what? Try 'get coin from ash'. Use 'iscan' to find interactable objects."); return }
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil { s.WriteLine("There is nowhere here to search."); return }
	raw := strings.TrimSpace(strings.Join(args, " "))
	target, source := raw, ""
	lower := strings.ToLower(raw)
	if index := strings.Index(lower, " from "); index >= 0 {
		target, source = strings.TrimSpace(raw[:index]), strings.TrimSpace(raw[index+6:])
	} else {
		// Also accept the compact MUD form: get coin ash.
		fields := strings.Fields(raw)
		for split := len(fields)-1; split > 0; split-- {
			candidateTarget := strings.Join(fields[:split], " ")
			candidateSource := strings.Join(fields[split:], " ")
			if findRoomObject(r, candidateTarget, candidateSource, s.Character) != nil {
				target, source = candidateTarget, candidateSource
				break
			}
		}
	}
	object := findRoomObject(r, target, source, s.Character)
	if object == nil && source != "" {
		container := findRoomObject(r, source, "", s.Character)
		if container != nil && container.Container && container.Open {
			key := containerStorageKey(r.ID, container.ID)
			storedItems := s.Character.ContainerStorage[key]
			for i, stored := range storedItems {
				if strings.Contains(strings.ToLower(stored.Name), strings.ToLower(target)) {
					s.Character.Inventory = append(s.Character.Inventory, stored)
					s.Character.ContainerStorage[key] = append(storedItems[:i], storedItems[i+1:]...)
					s.WriteLine("\x1b[1;32mYou take %s from %s.\x1b[0m", stored.Name, container.Name)
					s.persist()
					return
				}
			}
		}
	}
	if object == nil {
		if source != "" { s.WriteLine("You cannot find %s in %s here.", target, source) } else { s.WriteLine("You cannot find '%s' here. Use 'look' or 'iscan' to see what is present.", target) }
		return
	}
	if !object.Takeable {
		s.WriteLine("You cannot take %s; it is part of the surroundings.", object.Name)
		return
	}
	kind := "object"
	if object.ID == "scratched_bronze_coin" || object.ID == "harpy_letter" || object.ID == "charred_prayer_strip" { kind = "quest" }
	carried := item.Item{Name:object.Name, Description:object.Description, Tier:item.Common, Kind:kind}
	if object.StoredItem != nil { carried = *object.StoredItem }
	s.Character.Inventory = append(s.Character.Inventory, carried)
	s.Character.SetStoryFlag("object_taken_" + object.ID)
	s.WriteLine("\x1b[1;32mYou pick up %s.\x1b[0m", object.Name)
	if strings.EqualFold(object.ID, "scratched_bronze_coin") {
		s.Character.SetStoryFlag("found_scratched_coin")
		s.updateQuests()
		s.Character.Memories = append(s.Character.Memories, "I found a scratched bronze coin half-buried in cold campfire ash beneath Olympus.")
		s.WriteLine("The coin warms in your palm. For an instant, you smell salt water and hear an oar creak in the dark.")
	}
	s.persist()
}

func (s *Session) putItem(args []string) {
	if len(args) == 0 {
		s.WriteLine("Put what where? Try 'put bandage in chest'.")
		return
	}
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil {
		s.WriteLine("There is nowhere here to store an item.")
		return
	}
	raw := strings.TrimSpace(strings.Join(args, " "))
	lower := strings.ToLower(raw)
	separator := strings.Index(lower, " in ")
	if separator < 0 {
		separator = strings.Index(lower, " into ")
	}
	if separator < 0 {
		s.WriteLine("Name an open container: put <item> in <container>.")
		return
	}
	itemQuery := strings.TrimSpace(raw[:separator])
	containerQuery := strings.TrimSpace(raw[separator+4:])
	if strings.Contains(lower, " into ") && strings.Index(lower, " into ") == separator {
		containerQuery = strings.TrimSpace(raw[separator+6:])
	}
	container := findRoomObject(r, containerQuery, "", s.Character)
	if container == nil || !container.Container {
		s.WriteLine("You cannot find an interactable container matching '%s'.", containerQuery)
		return
	}
	if !container.Open {
		s.WriteLine("The %s is closed. Open it before putting anything inside.", container.Name)
		return
	}
	selected := -1
	for i, carried := range s.Character.Inventory {
		if strings.Contains(strings.ToLower(carried.Name), strings.ToLower(itemQuery)) {
			if selected >= 0 {
				s.WriteLine("Several carried items match '%s'. Use a more specific name.", itemQuery)
				return
			}
			selected = i
		}
	}
	if selected < 0 {
		s.WriteLine("You do not carry an item matching '%s'.", itemQuery)
		return
	}
	carried := s.Character.Inventory[selected]
	if carried.Relic || carried.Kind == "quest" {
		s.WriteLine("%s is tied to your soul or story and cannot be stored away.", carried.Name)
		return
	}
	if s.Character.ContainerStorage == nil {
		s.Character.ContainerStorage = map[string][]item.Item{}
	}
	key := containerStorageKey(r.ID, container.ID)
	s.Character.ContainerStorage[key] = append(s.Character.ContainerStorage[key], carried)
	s.Character.Inventory = append(s.Character.Inventory[:selected], s.Character.Inventory[selected+1:]...)
	s.WriteLine("You place %s inside %s.", carried.Name, container.Name)
	s.persist()
}

func (s *Session) dropItem(args []string) {
	if len(args) == 0 { s.WriteLine("Drop what? Use 'inventory' to see what you carry."); return }
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil { s.WriteLine("There is nowhere here to leave an item."); return }
	query := strings.ToLower(strings.Join(args, " "))
	selected := -1
	if number, err := strconv.Atoi(query); err == nil {
		if number >= 1 && number <= len(s.Character.Inventory) { selected = number-1 }
	} else {
		var matches []int
		for i, owned := range s.Character.Inventory {
			if strings.Contains(strings.ToLower(owned.Name), query) { matches = append(matches, i) }
		}
		if len(matches) == 1 { selected = matches[0] }
		if len(matches) > 1 {
			s.WriteLine("Several items match '%s'. Use the full name or inventory number:", query)
			for _, i := range matches { s.WriteLine("  %d) %s", i+1, s.Character.Inventory[i].Name) }
			return
		}
	}
	if selected < 0 { s.WriteLine("You do not carry an item matching '%s'.", query); return }
	owned := s.Character.Inventory[selected]
	if owned.Relic { s.WriteLine("%s is bound to your soul and cannot be dropped.", owned.Name); return }
	if owned.Kind == "quest" { s.WriteLine("%s is tied to your story and cannot be dropped.", owned.Name); return }
	if strings.EqualFold(owned.Name, s.Character.Weapon) || strings.EqualFold(owned.Name, s.Character.Armor) {
		s.WriteLine("You should unequip %s before dropping it.", owned.Name)
		return
	}
	id := fmt.Sprintf("dropped_%d_%s", time.Now().UnixNano(), strings.NewReplacer(" ", "_", "'", "").Replace(strings.ToLower(owned.Name)))
	description := owned.Description
	if description == "" { description = fmt.Sprintf("%s left behind by a traveler.", owned.Name) }
	object := world.NewObject(id, owned.Name, description, owned.Name, owned.Kind)
	stored := owned
	object.StoredItem = &stored
	object.Takeable = true
	r.Objects = append(r.Objects, object)
	s.Character.Inventory = append(s.Character.Inventory[:selected], s.Character.Inventory[selected+1:]...)
	s.WriteLine("You leave %s here.", owned.Name)
	s.persist()
}

func (s *Session) scanItems() {
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil { s.WriteLine("There is nowhere here to scan."); return }
	s.WriteLine("\x1b[1;36mINTERACTABLE SCAN — %s\x1b[0m", r.Name)
	found := false
	objects := visibleRoomObjects(r, s.Character)
	for i, object := range objects {
		found = true
		kind := "scenery"
		if object.Container { kind = "container" } else if object.Takeable { kind = "takeable" }
		if object.ContainerID != "" {
			if parent := findRoomObjectByID(r, object.ContainerID); parent != nil { s.WriteLine("  %d. %s [%s; in %s]", i+1, object.Name, kind, parent.Name); continue }
		}
		s.WriteLine("  %d. %s [%s]", i+1, object.Name, kind)
	}
	displayIndex := len(objects) + 1
	for _, container := range r.Objects {
		if container == nil || !container.Container || !container.Open || !container.VisibleIn(r.Objects) { continue }
		for _, stored := range s.Character.ContainerStorage[containerStorageKey(r.ID, container.ID)] {
			found = true
			s.WriteLine("  %d. %s [stored in %s; takeable]", displayIndex, stored.Name, container.Name)
			displayIndex++
		}
	}
	if !found { s.WriteLine("  Nothing obvious can be handled here. Look closely at the room description for clues.") }
	s.WriteLine("Try 'look <object>', 'examine <object>', or 'get <item> from <container>'.")
}

func (s *Session) scanTargets() {
	r := s.World.GetRoom(s.Character.RoomID)
	if r == nil { s.WriteLine("There is nowhere here to scan."); return }
	s.WriteLine("\x1b[1;36mTARGET SCAN — %s\x1b[0m", r.Name)
	npcs := s.World.NPCs[s.Character.RoomID]
	if len(npcs) > 0 {
		s.WriteLine("\x1b[1;37mPEOPLE\x1b[0m")
		for i, n := range npcs { s.WriteLine("  NPC %d: %s", i+1, n.Name) }
	}
	if s.Enemy == nil && s.EncounterClearedRoom != s.Character.RoomID { s.spawnEnemy() }
	if s.Enemy != nil && s.Enemy.HP > 0 {
		stance := "aggressive"
		if !s.EnemyAggressive {
			stance = strings.ToLower(s.Enemy.Behavior)
			if stance == "" { stance = "passive" }
		}
		s.WriteLine("\x1b[1;37mCREATURES\x1b[0m")
		s.WriteLine("  TARGET 1: %s — %s, HP %d/%d", s.Enemy.Name, stance, s.Enemy.HP, s.Enemy.MaxHP)
	}
	objects := visibleRoomObjects(r, s.Character)
	if len(objects) > 0 {
		s.WriteLine("\x1b[1;37mOBJECTS\x1b[0m")
		for i, object := range objects { s.WriteLine("  OBJECT %d: %s", i+1, object.Name) }
	}
	if len(npcs) == 0 && (s.Enemy == nil || s.Enemy.HP <= 0) && len(objects) == 0 {
		s.WriteLine("Nothing obvious stands out. Read the room description carefully.")
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
	s.EnemyAggressive = true
	if strings.HasPrefix(s.Character.RoomID, "tutorial_") {
		if s.Character.HasStoryFlag("tutorial_enemy_slain_" + s.Character.RoomID) { s.Enemy = nil; return }
		switch s.Character.RoomID {
		case "tutorial_yard":
			s.Enemy = &combat.Enemy{Name:"Training Shade", Description:"A small shadow woven from harmless dusk. It waits with its hands lowered; it will only defend itself if struck.", Level:1, HP:24, MaxHP:24, Defense:0, Damage:2, XP:0}
		case "tutorial_grove":
			s.Enemy = &combat.Enemy{Name:"Practice Satyr", Description:"A young satyr holds its staff loosely and watches you without hostility. It will not attack unless you attack first.", Level:1, HP:28, MaxHP:28, Defense:0, Damage:3, XP:0}
		default: return
		}
		s.EnemyAggressive = false
		s.WriteLine("\x1b[1;33mA %s waits nearby, but does not attack.\x1b[0m", s.Enemy.Name)
		s.WriteLine("%s", s.Enemy.Description)
		return
	}
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
		s.Enemy = &combat.Enemy{Name: "Street Shade", Behavior: "wary", Description: "A human-shaped shadow slips between the neon signs, moving against the light. It watches from the edge of the crowd and does not close in unless threatened.", Level: 2, HP: 80, MaxHP: 80, Defense: 5, Damage: 10, XP: 90}
	case "modern_metro":
		s.Enemy = &combat.Enemy{Name: "Echo Hound", Description: "A hound made of static and old memories emerges from the empty train.", Level: 4, HP: 145, MaxHP: 145, Defense: 7, Damage: 15, XP: 170}
	case "modern_styx":
		s.Enemy = &combat.Enemy{Name: "Styx Wraith", Description: "Black water rises into the shape of a veiled figure, carrying coins from lives you never lived.", Level: 5, HP: 185, MaxHP: 185, Defense: 8, Damage: 18, XP: 230}
	case "modern_sanctum":
		s.Enemy = &combat.Enemy{Name: "Bronze Sentinel", Description: "A bronze guardian unfolds from the hidden temple wall, its eyes burning with borrowed starlight.", Level: 6, HP: 220, MaxHP: 220, Defense: 9, Damage: 20, XP: 280}
	case "future_city":
		s.Enemy = &combat.Enemy{Name: "Chronal Warden", Behavior: "wary", Description: "A guardian of fractured seconds steps from a ripple in the air, studying you as if comparing several possible futures.", Level: 2, HP: 95, MaxHP: 95, Defense: 5, Damage: 11, XP: 100}
	case "future_skyway":
		s.Enemy = &combat.Enemy{Name: "Storm Automaton", Description: "A machine of celestial bronze and lightning blocks the road above the clouds.", Level: 4, HP: 150, MaxHP: 150, Defense: 7, Damage: 15, XP: 180}
	case "future_moon":
		s.Enemy = &combat.Enemy{Name: "Moonshade", Behavior: "wary", Description: "A pale shadow detaches itself from the lunar sanctuary's wall and reaches for your memories, but pauses when it recognizes the thread around your wrist.", Level: 4, HP: 145, MaxHP: 145, Defense: 7, Damage: 14, XP: 180}
	case "far_era":
		s.Enemy = &combat.Enemy{Name: "Last Shore Titan", Description: "The drowned world's final guardian rises from the surf, carrying the weight of vanished centuries.", Level: 5, HP: 180, MaxHP: 180, Defense: 8, Damage: 18, XP: 220}
	default:
		return
	}
	behavior := strings.ToLower(s.Enemy.Behavior)
	if behavior == "" { behavior = "aggressive" }
	s.EnemyAggressive = behavior == "aggressive"
	switch behavior {
	case "passive":
		s.WriteLine("\x1b[1;32m%s is here, watching you without hostility.\x1b[0m", s.Enemy.Name)
	case "wary":
		s.WriteLine("\x1b[1;33m%s circles at a distance, wary but not yet attacking.\x1b[0m", s.Enemy.Name)
	default:
		s.WriteLine("\x1b[1;31mA %s appears, ready to attack!\x1b[0m", s.Enemy.Name)
	}
	s.WriteLine("%s",s.Enemy.Description)
}

func (s *Session) canEnter(id string) bool {
	c:=s.Character
	if strings.HasPrefix(id, "tutorial_") && (!c.HasStoryFlag("tutorial_available") || c.HasStoryFlag("tutorial_complete")) {
		s.WriteLine("The tutorial threshold has closed. That first passage cannot be revisited.")
		return false
	}
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
	if roomID == "therons_forge" { s.Character.SetStoryFlag("forge_discovered") }
	if strings.HasPrefix(roomID, "tutorial_") { return }
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
	// The prologue is a safe place to experiment, not a command lockstep.
	// Players may explore in their own order and can leave through the gate
	// whenever they are ready; tutorial hints remain available on request.
	r:=s.World.GetRoom(s.Character.RoomID)
	if r==nil{return}
	next,ok:=r.Exits[direction]
	if !ok {s.WriteLine("You cannot go that way.");return}
	if s.Enemy!=nil && s.Enemy.HP>0 {s.WriteLine("You cannot leave while the %s still stands.",s.Enemy.Name);return}
	if !s.canEnter(next) { return }
	if r.ID == "village_lane" && direction == "north" && s.TutorialStep < 9 {
		s.WriteLine("Damon raises a hand. Finish the village lessons first: inspect and speak with him, use 'hint', then practice 'exits' and 'map'.")
		return
	}
	s.Character.RoomID = next
	s.EncounterClearedRoom = ""
	s.PendingEquip = nil
	s.recordDiscovery(next)
	s.WriteLine("You travel %s.", direction)
	if s.AutoMap { s.localMap() }
	s.updateQuests()
	s.look()
	// Aggressive world encounters get an opening action as soon as they engage.
	// Tutorial creatures explicitly set EnemyAggressive=false and never ambush.
	if s.Enemy != nil && s.Enemy.HP > 0 && s.EnemyAggressive {
		s.WriteLine("The %s lunges before you can settle into the room.", s.Enemy.Name)
		s.enemyTurn()
	}
	if s.Character.RoomID == "village_square" && r.ID == "tutorial_gate" && s.Character.HasStoryFlag("tutorial_available") {
		s.Character.SetStoryFlag("tutorial_complete")
		s.Character.SetStoryFlag("tutorial_finished")
		s.Character.StoryFlags["tutorial_available"] = false
		s.PrologueStep = 39
		s.WriteLine("\x1b[1;32mTHE FIRST THREAD — COMPLETE\x1b[0m")
		s.WriteLine("The marble gate seals behind you. Your training is remembered, but the tutorial itself is no longer reachable.")
		s.WriteLine("You have entered Fatewalker proper. Use 'help list', 'journal', 'quests', and 'look' as you explore.")
	} else if s.Character.RoomID == "village_lane" && s.TutorialStep == 1 {
		s.advanceTutorial(2)
	} else if s.Character.RoomID == "olympus_foothills" && s.TutorialStep == 9 {
		s.advanceTutorial(10)
	}
}

func (s *Session) attack(args []string) {
	if s.Enemy == nil || s.Enemy.HP <= 0 {
		s.WriteLine("There is nothing here to fight.")
		return
	}
	if s.Character.HasStoryFlag("tutorial_available") {
		if s.Character.RoomID == "tutorial_yard" && s.PrologueStep < 12 {
			s.WriteLine("Inspect the Training Shade first with 'look Training Shade'.")
			return
		}
		if s.Character.RoomID == "tutorial_grove" && s.PrologueStep < 20 {
			s.WriteLine("Inspect the Practice Satyr first with 'look Satyr'.")
			return
		}
	}
	if s.TutorialStep == 11 && s.Enemy.Name == "Harpy" && !s.Character.HasStoryFlag("tutorial_enemy_inspected") {
		s.WriteLine("Before striking, inspect the creature with 'look Harpy'. Learn what you are facing.")
		return
	}
	if len(args)>0{target:=strings.Join(args," ");if target!="1"&&!strings.Contains(strings.ToLower(s.Enemy.Name),target){s.WriteLine("Your current target is %s. Use 'attack 1' or 'attack %s'.",s.Enemy.Name,strings.ToLower(s.Enemy.Name));return}}
	s.EnemyAggressive = true
	s.basicAttackRound()
}


func (s *Session) initializeCombatMode() {
	if s.Character == nil { return }
	// Characters created before this preference existed default to automatic
	// combat the first time they are loaded after this update.
	if !s.Character.CombatModeSet {
		s.Character.AutoCombat = true
		s.Character.CombatModeSet = true
	}
}

func (s *Session) combatMode(args []string, short bool) {
	if s.Character == nil { return }
	if len(args) == 0 {
		if short {
			s.Character.AutoCombat = !s.Character.AutoCombat
			s.Character.CombatModeSet = true
		} else {
			s.WriteLine("Combat mode is %s. Use 'combatmode on/off' or 'pc' to toggle.", combatModeName(s.Character.AutoCombat))
			return
		}
	} else {
		switch strings.ToLower(args[0]) {
		case "on", "auto", "automatic":
			s.Character.AutoCombat = true
		case "off", "manual", "nonautomatic", "non-automatic":
			s.Character.AutoCombat = false
		case "toggle":
			s.Character.AutoCombat = !s.Character.AutoCombat
		case "status", "show":
			s.WriteLine("Combat mode is %s.", combatModeName(s.Character.AutoCombat))
			return
		default:
			s.WriteLine("Use 'pc' to toggle, 'pc on/off', or 'combatmode on/off/status'.")
			return
		}
		s.Character.CombatModeSet = true
	}
	s.WriteLine("Combat mode: %s. %s", combatModeName(s.Character.AutoCombat), map[bool]string{true:"Your attacks continue automatically after you engage an enemy.", false:"You must type attack for each basic attack."}[s.Character.AutoCombat])
	s.persist()
}

func combatModeName(automatic bool) string {
	if automatic { return "AUTOMATIC" }
	return "MANUAL"
}

// automaticCombatRound is called by the live combat ticker while the player
// is idle. It stops naturally when the enemy is defeated or the player falls.
func (s *Session) automaticCombatRound() {
	if s.Character == nil || !s.Character.AutoCombat || s.Enemy == nil || s.Enemy.HP <= 0 || !s.EnemyAggressive { return }
	s.basicAttackRound()
}

// basicAttackRound resolves one player weapon swing and the enemy's response.
func (s *Session) basicAttackRound() {
	if s.Enemy == nil || s.Enemy.HP <= 0 { return }
	w := s.currentWeapon()
	result := combat.Attack(s.Character.Name, w, s.Enemy, s.Character.AttackPower())
	if result.Damage > 0 {
		totalDamage, bonus := s.applySoulRelicStrike(result.Damage)
		if bonus > 0 {
			s.Enemy.HP -= bonus
			if s.Enemy.HP < 0 { s.Enemy.HP = 0 }
			result.Damage = totalDamage
			result.Killed = s.Enemy.HP == 0
			result.Text += fmt.Sprintf("\\nThe %s answers your strike, adding %d damage. (%d total damage)", s.soulRelicStrikeName(), bonus, totalDamage)
			if result.Killed { result.Text += " The " + s.Enemy.Name + " collapses." }
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
	if len(args)==0 { s.WriteLine("Cast which power? Try 'powers'."); return }
	powerName:=strings.Join(args," ")
	// A final argument matching the current enemy is treated as an optional target.
	if s.Enemy != nil && len(args)>1 {
		target:=strings.ToLower(args[len(args)-1])
		if target=="1" || strings.Contains(strings.ToLower(s.Enemy.Name),target) {
			powerName=strings.Join(args[:len(args)-1]," ")
		} else if strings.Contains(strings.ToLower(strings.Join(args," ")),strings.ToLower(s.Enemy.Name)) {
			powerName=strings.Replace(strings.ToLower(strings.Join(args," ")),strings.ToLower(s.Enemy.Name),"",1)
			powerName=strings.TrimSpace(powerName)
		}
	}
	s.castPower(powerName)
}

func (s *Session) castPower(powerName string) {
	if s.Enemy==nil || s.Enemy.HP<=0 {s.WriteLine("There is no enemy to target.");return}
	if s.Character.Domain=="" {s.WriteLine("You have no awakened divine domain. Try 'awaken storm'.");return}
	if powerName=="" {s.WriteLine("Cast which power? Try 'powers'.");return}
	var chosen *power.Power
	for _,p:=range power.Unlocked(s.Character.Domain,s.Character.Level) {
		if strings.EqualFold(p.Name,powerName) || strings.EqualFold(strings.ReplaceAll(p.Name," ",""),strings.ReplaceAll(powerName," ","")) {q:=p;chosen=&q;break}
	}
	if chosen==nil {s.WriteLine("That power is not yet awakened.");return}
	if s.Character.Mana<chosen.ManaCost {s.WriteLine("Your divine reserves are exhausted.");return}
	s.Character.Mana-=chosen.ManaCost
	result:=combat.Cast(s.Character.Name,chosen.Name,chosen.Damage,string(chosen.Domain),s.Enemy)
	s.WriteLine("\x1b[1;35m%s\x1b[0m",result.Text)
	if result.Killed { s.defeatEnemy(true); return }
	s.enemyTurn()
}

func (s *Session) quicklist(args []string) {
	if s.Character == nil { return }
	if len(args)==0 {
		s.WriteLine("COMBAT QUICKLIST")
		for i:=0;i<len(s.Character.Quicklist);i++ {
			name:=s.Character.Quicklist[i]
			if name=="" { name="(empty)" }
			s.WriteLine("  ql %d <target>  — %s",i+1,name)
		}
		s.WriteLine("Set: ql set <slot> <power name> | Clear: ql clear <slot>")
		return
	}
	if strings.EqualFold(args[0],"set") {
		if len(args)<3 { s.WriteLine("Use: ql set <slot> <power name>"); return }
		slot,err:=strconv.Atoi(args[1])
		if err!=nil || slot<1 || slot>len(s.Character.Quicklist) { s.WriteLine("Choose a quicklist slot from 1 to %d.",len(s.Character.Quicklist)); return }
		name:=strings.Join(args[2:]," ")
		if s.Character.Domain=="" { s.WriteLine("Choose a divine domain first with 'awaken'."); return }
		valid:=false
		for _,p:=range power.Unlocked(s.Character.Domain,s.Character.Level) {
			if strings.EqualFold(p.Name,name) || strings.EqualFold(strings.ReplaceAll(p.Name," ",""),strings.ReplaceAll(name," ","")) { name=p.Name; valid=true; break }
		}
		if !valid { s.WriteLine("That power is not currently unlocked. Check 'powers'."); return }
		s.Character.Quicklist[slot-1]=name
		s.WriteLine("Quicklist slot %d set to %s.",slot,name)
		s.persist()
		return
	}
	if strings.EqualFold(args[0],"clear") {
		if len(args)!=2 { s.WriteLine("Use: ql clear <slot>"); return }
		slot,err:=strconv.Atoi(args[1])
		if err!=nil || slot<1 || slot>len(s.Character.Quicklist) { s.WriteLine("Choose a quicklist slot from 1 to %d.",len(s.Character.Quicklist)); return }
		s.Character.Quicklist[slot-1]=""
		s.WriteLine("Quicklist slot %d cleared.",slot)
		s.persist()
		return
	}
	slot,err:=strconv.Atoi(args[0])
	if err!=nil || slot<1 || slot>len(s.Character.Quicklist) { s.WriteLine("Use 'ql' to list slots, 'ql set <slot> <power>', or 'ql <slot> <target>'."); return }
	name:=s.Character.Quicklist[slot-1]
	if name=="" { s.WriteLine("Quicklist slot %d is empty. Set it with 'ql set %d <power name>'.",slot,slot); return }
	if len(args)>1 && s.Enemy!=nil {
		target:=strings.ToLower(strings.Join(args[1:]," "))
		if target!="1" && !strings.Contains(strings.ToLower(s.Enemy.Name),target) {
			s.WriteLine("Your current target is %s. Use 'ql %d %s'.",s.Enemy.Name,slot,strings.ToLower(s.Enemy.Name))
			return
		}
	}
	s.castPower(name)
}

func (s *Session) craft(args []string) {
	if s.Character.RoomID != "therons_forge" {
		s.WriteLine("Crafting requires Theron's forge. Travel east from the Olympus foothills.")
		return
	}
	if len(args) == 0 {
		s.WriteLine("FORGE RECIPES")
		s.WriteLine("  craft stormguard cuirass — 1 Harpy Feather + 60 gold; rare armor (+7)")
		s.WriteLine("  craft laurel spear      — 1 Satyr Horn + 55 gold; rare spear (23 damage)")
		s.WriteLine("  craft manticore mantle  — 1 Manticore Spine + 90 gold; epic armor (+10)")
		return
	}
	name := strings.ToLower(strings.Join(args, " "))
	type recipe struct {
		name, material string
		cost int
		output item.Item
	}
	recipes := []recipe{
		{name:"stormguard cuirass", material:"harpy feather", cost:60, output:item.Item{Name:"stormguard cuirass", Tier:item.Rare, Armor:7, Kind:"armor"}},
		{name:"laurel spear", material:"satyr horn", cost:55, output:item.Item{Name:"laurel spear", Tier:item.Rare, Damage:23, Kind:"spear"}},
		{name:"manticore mantle", material:"manticore spine", cost:90, output:item.Item{Name:"manticore mantle", Tier:item.Epic, Armor:10, Kind:"armor"}},
	}
	var chosen *recipe
	for i := range recipes {
		if strings.EqualFold(recipes[i].name, name) || strings.Contains(recipes[i].name, name) {
			if chosen != nil {
				s.WriteLine("That matches more than one recipe. Use the full recipe name.")
				return
			}
			chosen = &recipes[i]
		}
	}
	if chosen == nil {
		s.WriteLine("Theron does not know that recipe. Type 'craft' to list available recipes.")
		return
	}
	materialIndex := -1
	for i, it := range s.Character.Inventory {
		if strings.EqualFold(it.Name, chosen.material) {
			materialIndex = i
			break
		}
	}
	if materialIndex < 0 {
		s.WriteLine("You need 1 %s before Theron can make that.", chosen.material)
		return
	}
	if s.Character.Gold < chosen.cost {
		s.WriteLine("You need %d gold; you currently have %d.", chosen.cost, s.Character.Gold)
		return
	}
	s.Character.Gold -= chosen.cost
	s.Character.Inventory = append(s.Character.Inventory[:materialIndex], s.Character.Inventory[materialIndex+1:]...)
	s.Character.Inventory = append(s.Character.Inventory, chosen.output)
	s.WriteLine("Theron works the material over the anvil. Sparks leap like tiny stars.")
	s.WriteLine("Crafted %s (%s). Gold remaining: %d.", chosen.output.Name, chosen.output.TierName(), s.Character.Gold)
	s.Character.SetStoryFlag("crafted_" + strings.ReplaceAll(chosen.output.Name, " ", "_"))
	s.Character.SetStoryFlag("crafted_any_gear")
	s.updateQuests()
	s.Character.SetStoryFlag("crafted_any_gear")
	s.updateQuests()
}

func (s *Session) defeatEnemy(divine bool) {
	if s.Enemy == nil {
		return
	}
	enemy := s.Enemy
	if strings.HasPrefix(s.Character.RoomID, "tutorial_") {
		s.Character.SetStoryFlag("tutorial_enemy_slain_" + s.Character.RoomID)
	}
	s.advanceQuestKill(enemy.Name)
	s.awardExperience(enemy.XP)
	var material item.Item
	switch enemy.Name {
	case "Harpy":
		material = item.Item{Name:"harpy feather", Tier:item.Common, Kind:"material"}
	case "Satyr":
		material = item.Item{Name:"satyr horn", Tier:item.Common, Kind:"material"}
	case "Manticore":
		material = item.Item{Name:"manticore spine", Tier:item.Uncommon, Kind:"material"}
	}
	if material.Name != "" {
		s.Character.Inventory = append(s.Character.Inventory, material)
		s.WriteLine("Harvested crafting material: %s.", material.Name)
	}
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
	s.EnemyAggressive = false
	s.RelicWardSpent = false
	s.RelicStrikeSpent = false
	if s.Character.RoomID == "olympus_foothills" && s.TutorialStep == 11 && enemy.Name == "Harpy" {
		s.advanceTutorial(12)
	}
	if s.Character.RoomID == "tutorial_grove" && enemy.Name == "Practice Satyr" {
		s.WriteLine("The satyr lowers its staff and steps aside. There is no road north through the grove; the familiar courtyard lies south.")
		s.WriteLine("From the courtyard, the storehouse is east, and the marble gate beyond it leads back into the wider world. You can take that route whenever you are ready.")
	}
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

func (s *Session) guard() {
	if s.Enemy == nil || s.Enemy.HP <= 0 || !s.EnemyAggressive {
		s.WriteLine("There is no aggressive enemy to defend against.")
		return
	}
	s.Guarding = true
	s.WriteLine("You brace behind your guard, watching the enemy's shoulders and waiting for the strike.")
	s.enemyTurn()
}

func (s *Session) enemyTurn() {
	if s.Enemy==nil || s.Enemy.HP<=0 || !s.EnemyAggressive{return}
	result:=combat.EnemyAttack(s.Enemy,s.Character.DefensePower())
	if s.Guarding {
		if result.Damage > 0 {
			before := result.Damage
			result.Damage = (result.Damage + 1) / 2
			result.Text += fmt.Sprintf("\nYou hold your guard and reduce the blow from %d to %d damage.", before, result.Damage)
		} else {
			result.Text += "\nYour guard holds; the attack finds no opening."
		}
		s.Guarding = false
	}
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

func (s *Session) useItem(args []string) {
	if len(args) == 0 {
		s.WriteLine("Use what? Try 'inv' to inspect your pack, then 'use <item name>'.")
		return
	}
	query := strings.ToLower(strings.Join(args, " "))
	selected := -1
	var matches []int
	for i, it := range s.Character.Inventory {
		if it.Healing <= 0 && it.ManaRestore <= 0 { continue }
		name := strings.ToLower(it.Name)
		if name == query { selected = i; break }
		if strings.Contains(name, query) { matches = append(matches, i) }
	}
	if selected < 0 && len(matches) == 1 { selected = matches[0] }
	if selected < 0 && len(matches) > 1 {
		s.WriteLine("Several consumables match '%s'. Choose one by its full name:", query)
		for _, i := range matches { s.WriteLine("  %d) %s", i+1, s.Character.Inventory[i].Name) }
		return
	}
	if selected < 0 {
		s.WriteLine("You do not carry a usable consumable matching '%s'. Check 'inv' or visit a merchant.", query)
		return
	}
	it := s.Character.Inventory[selected]
	healed, restored := 0, 0
	if it.Healing > 0 {
		healed = it.Healing
		if s.Character.HP+healed > s.Character.MaxHP { healed = s.Character.MaxHP-s.Character.HP }
		if healed < 0 { healed = 0 }
		s.Character.HP += healed
	}
	if it.ManaRestore > 0 {
		restored = it.ManaRestore
		if s.Character.Mana+restored > s.Character.MaxMana { restored = s.Character.MaxMana-s.Character.Mana }
		if restored < 0 { restored = 0 }
		s.Character.Mana += restored
	}
	if healed == 0 && restored == 0 {
		s.WriteLine("%s would have no effect right now.", it.Name)
		return
	}
	s.Character.Inventory = append(s.Character.Inventory[:selected], s.Character.Inventory[selected+1:]...)
	s.WriteLine("You use %s: restored %d health and %d mana. The item is consumed.", it.Name, healed, restored)
	if s.Enemy != nil && s.Enemy.HP > 0 && s.EnemyAggressive { s.enemyTurn() }
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
	if s.Enemy == nil || s.Enemy.HP <= 0 {
		s.WriteLine("You are not in combat.")
		return
	}
	if s.Character.HasStoryFlag("tutorial_available") && strings.HasPrefix(s.Character.RoomID, "tutorial_") {
		s.WriteLine("This is a controlled practice encounter. Stay with the lesson and defeat the %s to continue.", s.Enemy.Name)
		return
	}
	name := s.Enemy.Name
	s.WriteLine("You turn to break away from the %s. It gets one last chance to strike.", name)
	s.enemyTurn()
	if s.Enemy == nil {
		// The enemy's parting attack defeated the character; recovery already ran.
		return
	}
	s.EncounterClearedRoom = s.Character.RoomID
	s.Enemy = nil
	s.Guarding = false
	s.RelicWardSpent = false
	s.RelicStrikeSpent = false
	s.WriteLine("You escape the %s and retreat. You gain no victory reward; survival is the prize.", name)
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
		if it.Kind == "material" { tags += "  \x1b[1;33m[MATERIAL]\x1b[0m" }
		s.WriteLine("  %2d. %s%s%s%s", i+1, s.color(it.TierColor()), it.Name, s.color("0"), tags)
		details := fmt.Sprintf("%s  |  Armor +%d  |  Damage +%d", it.TierName(), it.Armor, it.Damage)
		if it.Healing > 0 { details += fmt.Sprintf("  |  Heals %d", it.Healing) }
		if it.ManaRestore > 0 { details += fmt.Sprintf("  |  Restores %d mana", it.ManaRestore) }
		s.WriteLine("      %s", details)
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
		s.WriteLine("  %s — level %d — %s — %d mana — %d base damage",p.Name,p.UnlockLevel,status,p.ManaCost,p.Damage)
		s.WriteLine("    %s",p.Description)
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
	bar := func(current, maximum, width int) string {
		if maximum < 1 { maximum = 1 }
		if current < 0 { current = 0 }
		if current > maximum { current = maximum }
		filled := current * width / maximum
		return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	}
	title := strings.ToUpper(c.Name)
	s.WriteLine("")
	s.WriteLine("\x1b[1;36m╭────────────────────────────── FATEWALKER ──────────────────────────────╮\x1b[0m")
	s.WriteLine("\x1b[1;36m│\x1b[0m \x1b[1;37m%-70s\x1b[0m \x1b[1;36m│\x1b[0m", compactText(title,70))
	s.WriteLine("\x1b[1;36m╰────────────────────────────────────────────────────────────────────────╯\x1b[0m")
	s.WriteLine("  \x1b[1;33mLIFE %d\x1b[0m   Level %d   %s era   Rebirths %d", c.Life, c.Level, strings.Title(c.Era), c.Rebirths)
	domain := "Unawakened"
	if c.Domain != "" { domain = strings.Title(string(c.Domain)) }
	s.WriteLine("  Divine path: %-18s  Divinity: %-3d  Gift: %s", domain, c.Divinity, c.LifeGift)
	s.WriteLine("")
	s.WriteLine("\x1b[1;37mVITALS\x1b[0m")
	s.WriteLine("  HP    \x1b[1;32m%s\x1b[0m %d/%d", bar(c.HP,c.MaxHP,22), c.HP,c.MaxHP)
	s.WriteLine("  Mana  \x1b[1;36m%s\x1b[0m %d/%d", bar(c.Mana,c.MaxMana,22), c.Mana,c.MaxMana)
	s.WriteLine("  XP    %d   •   %d to next level", c.Experience, progression.XPToNextLevel(c.Level,c.Experience))
	s.WriteLine("")
	s.WriteLine("\x1b[1;37mCOMBAT & EQUIPMENT\x1b[0m")
	s.WriteLine("  Attack %-4d Defense %-4d Armor %-4d Gold %d", c.AttackPower(), c.DefensePower(), c.ArmorPower(), c.Gold)
	s.WriteLine("  Weapon: %s", equippedName(c.Weapon, c.Inventory))
	s.WriteLine("  Armor:  %s", equippedName(c.Armor, c.Inventory))
	s.WriteLine("  Pack: %d items   Soul relics: %d   Combat: %s", len(c.Inventory), soulRelicCount(c.Inventory), combatModeName(c.AutoCombat))
	s.WriteLine("")
	s.WriteLine("\x1b[1;37mQUICKLIST\x1b[0m")
	for i, powerName := range c.Quicklist {
		if powerName == "" { powerName = "—" }
		s.WriteLine("  %d. %s", i+1, powerName)
	}
	s.WriteLine("")
	s.WriteLine("\x1b[1;37mSOUL LEGACY\x1b[0m")
	s.WriteLine("  Memories %d  •  Scars %d  •  Oaths %d", len(c.Memories), len(c.Scars), len(c.Oaths))
	s.WriteLine("  Favors %d  •  Curses %d  •  Echoes %d", len(c.Favors), len(c.Curses), len(c.Echoes))
	s.WriteLine("\x1b[90m  Use 'inventory', 'powers', 'quests', and 'soul' for the full records.\x1b[0m")
}

func equippedName(name string, inventory []item.Item) string {
	for _, owned := range inventory {
		if strings.EqualFold(owned.Name, name) { return owned.Name }
	}
	if name == "" { return "None" }
	return name + " (not in pack)"
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

func (s *Session) currentNPC(index int) *world.NPC {
	npcs := s.World.NPCs[s.Character.RoomID]
	if index < 1 || index > len(npcs) { return nil }
	return npcs[index-1]
}

// parseNPCSelection accepts a numeric selector or an NPC name at the start
// of the arguments, leaving any remaining words as the conversation topic.
func (s *Session) parseNPCSelection(args []string) (int, []string) {
	if len(args) == 0 { return 1, args }
	if n, err := strconv.Atoi(args[0]); err == nil { return n, args[1:] }
	npcs := s.World.NPCs[s.Character.RoomID]
	for index, npc := range npcs {
		nameParts := strings.Fields(strings.ToLower(npc.Name))
		for i := range nameParts {
			nameParts[i] = strings.Trim(nameParts[i], ".,!?;:\\\"'()[]{}")
		}
		if len(nameParts) == 0 { continue }
		// Accept the full display name or the NPC's first-name alias.
		for _, nameLength := range []int{len(nameParts), 1} {
			if len(args) < nameLength { continue }
			matches := true
			for i := 0; i < nameLength; i++ {
				if !strings.EqualFold(strings.Trim(args[i], ".,!?;:\\\"'()[]{}"), nameParts[i]) { matches = false; break }
			}
			if matches { return index + 1, args[nameLength:] }
		}
	}
	return 1, args
}

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

func containsTopicKeyword(topic string, keywords ...string) bool {
	fields := strings.Fields(strings.ToLower(topic))
	for i := range fields { fields[i] = strings.Trim(fields[i], ".,!?;:\\\"'()[]{}") }
	normalized := " " + strings.Join(fields, " ") + " "
	for _, keyword := range keywords {
		keywordFields := strings.Fields(strings.ToLower(keyword))
		for i := range keywordFields { keywordFields[i] = strings.Trim(keywordFields[i], ".,!?;:\\\"'()[]{}") }
		needle := " " + strings.Join(keywordFields, " ") + " "
		if needle != "  " && strings.Contains(normalized, needle) { return true }
	}
	return false
}

func(s *Session) talk(args []string) {
 if s.Enemy!=nil&&s.Enemy.HP>0{s.WriteLine("The %s keeps you too busy to talk.",s.Enemy.Name);return}
 index,topicArgs:=s.parseNPCSelection(args);n:=s.currentNPC(index)
 if n==nil{if len(s.World.NPCs[s.Character.RoomID])==0{s.WriteLine("There is no one here willing to speak with you.")}else{s.WriteLine("Choose a valid NPC number from 'look'.")};return}
 topic:="hello";if len(topicArgs)>0{topic=strings.Join(topicArgs," ")}
 matchedTopic, dialogueText := n.MatchDialogue(topic)
 // Special dialogue actions use the same keyword intent matching as regular
 // dialogue, so a full question like "what choices do I have?" still works.
 if matchedTopic == "" && n.ID == "pythia" && containsTopicKeyword(topic, "choice", "choices") {
	matchedTopic = "choices"
 }
 if matchedTopic != "" { topic = matchedTopic }
 if greeting:=s.persistentNPCGreeting(n,topic);greeting!=""{s.WriteLine("%s",greeting);return}
 if topic=="hello"||topic=="greeting"{if greeting:=s.factionGreeting(n);greeting!=""{s.WriteLine("%s",greeting);return}}
 if n.ID=="pythia"&&(topic=="hello"||topic=="greeting"){switch{case s.Character.HasStoryFlag("oracle_trust"):s.WriteLine("\x1b[1;36mPythia smiles faintly. \"You kept the thread I gave you. The river will test that promise when you least expect it.\"\x1b[0m");return;case s.Character.HasStoryFlag("oracle_defied"):s.WriteLine("\x1b[1;36mPythia regards you without anger. \"Still walking your own road, I see. Even the Fates have learned to leave a little room for defiance.\"\x1b[0m");return}}
 if topic=="choice"||topic=="choices"{if n.ID=="pythia"{if s.Character.HasStoryFlag("oracle_choice_made"){s.WriteLine("Pythia studies you. \"The river has recorded your answer. You cannot make that choice unmade.\"");return};s.WriteLine("Pythia's voice falls to a whisper: \"When the Fates offer a thread, will you trust the pattern or cut your own path?\"");s.WriteLine("  choose trust — accept the Oracle's guidance and swear to remember it.");s.WriteLine("  choose defy  — reject prophecy and bear the consequences alone.");return}}
 if n.ID == "village_guide" && s.TutorialStep == 4 &&
		(topic == "road" || topic == "mountain" || topic == "north") &&
		!s.Character.HasStoryFlag("tutorial_hint_used") {
		s.WriteLine("Damon taps the milestone. \"Before you ask about the road, try 'hint' and see what guidance is available.\"")
		return
	}
 if t:=dialogueText;t!="" {
		s.WriteLine("\x1b[1;36m%s:\x1b[0m %s",n.Name,t)
		// These story flags turn optional conversations into a small, persistent
		// quest chain without making the main route depend on it.
		if n.ID == "charon" && (matchedTopic == "memory" || matchedTopic == "remember" || matchedTopic == "rebirth") {
			if !s.Character.HasStoryFlag("charon_memory_heard") {
				s.Character.SetStoryFlag("charon_memory_heard")
				s.WriteLine("\x1b[1;33mA new story thread has appeared: A Memory for the Dead.\x1b[0m")
			}
		}
		if n.ID == "village_guide" && s.TutorialStep == 3 && (topic == "hello" || topic == "greeting") {
			s.advanceTutorial(4)
		} else if n.ID == "village_guide" && s.TutorialStep == 4 && (topic == "village" || topic == "home") {
			s.advanceTutorial(5)
		} else if n.ID == "village_guide" && s.TutorialStep == 6 && (topic == "road" || topic == "mountain" || topic == "north") && s.Character.HasStoryFlag("tutorial_hint_used") {
			s.advanceTutorial(7)
		}
		s.updateQuests()
		return
	}
	s.WriteLine("\x1b[1;36m%s:\x1b[0m \"Ask me about the things that matter here.\"",n.Name)
}

func (s *Session) persistentNPCGreeting(n *world.NPC, topic string) string {
	if n == nil || n.ID != "athens_vendor" {
		return ""
	}
	if containsTopicKeyword(topic, "museum", "artifact", "artifacts", "replica", "replicas") && s.Character.HasStoryFlag("museum_echo_uncovered") {
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


// offerMemory lets the player contribute an actual remembered experience to
// the Underworld quest. The text is stored in the character's soul record;
// merely mentioning the word "memory" to the shade is not enough to complete it.
func (s *Session) offer(args []string) {
	if len(args) > 0 && strings.EqualFold(args[0], "memory") {
		s.offerMemory(args)
		return
	}
	if len(args) < 3 {
		s.WriteLine("Offer what, and to whom? Try 'offer coin to bowl' or 'offer memory <specific remembered moment>'.")
		return
	}
	joined := strings.Join(args, " ")
	lower := strings.ToLower(joined)
	to := strings.Index(lower, " to ")
	if to < 0 {
		s.WriteLine("Name the recipient: offer <item> to <person or object>.")
		return
	}
	itemName := strings.TrimSpace(joined[:to])
	recipient := stripLeadingArticle(strings.TrimSpace(joined[to+4:]))
	s.offerItem(itemName, recipient)
}

func (s *Session) offerItem(itemName, recipient string) {
	if s.Enemy != nil && s.Enemy.HP > 0 {
		s.WriteLine("The %s leaves you no room to make an offering.", s.Enemy.Name)
		return
	}
	if s.Character.RoomID == "fields_asphodel" && (strings.Contains(recipient, "shade") || strings.Contains(recipient, "asphodel")) {
		if s.Character.HasStoryFlag("offered_harpy_letter_to_shade") {
			s.WriteLine(`The shade rests one hand over the place where the letter was kept. "I have the witness now. I will not ask you to surrender it twice."`)
			return
		}
		selected := -1
		for i, carried := range s.Character.Inventory {
			if strings.Contains(strings.ToLower(carried.Name), "feather-folded letter") && strings.Contains(strings.ToLower(carried.Name), strings.ToLower(itemName)) {
				selected = i
				break
			}
		}
		if selected < 0 {
			s.WriteLine("You need to be carrying the feather-folded letter. The shade cannot receive a story from an empty hand.")
			return
		}
		if !s.Character.HasStoryFlag("read_harpy_letter") {
			s.WriteLine(`The shade studies the paper from a distance. "A mark without its words is only a shape. Read the letter first, and decide what you are carrying here."`)
			return
		}
		present := false
		for _, npc := range s.World.NPCs[s.Character.RoomID] {
			if npc.ID == "asphodel_shade" { present = true; break }
		}
		if !present {
			s.WriteLine("There is no shade here to receive the letter.")
			return
		}
		s.Character.Inventory = append(s.Character.Inventory[:selected], s.Character.Inventory[selected+1:]...)
		s.Character.SetStoryFlag("offered_harpy_letter_to_shade")
		s.Character.Memories = append(s.Character.Memories, "I gave the feather-folded letter to the Unremembered Shade. It recognized the three-line witness mark but could not recover the writer's name.")
		s.WriteLine("The shade takes the letter with both hands. Its fingers stop at the three-line mark.")
		s.WriteLine(`"I know this mark," it whispers. "Or I know the space it leaves behind. Someone stood in a doorway so another could pass. I cannot tell you who they were, and I will not pretend that recognition is the same as remembering. But this is a witness. Let it stay here, where the forgotten are still allowed to matter."`)
		s.WriteLine("The shade folds the letter carefully and places it beside the clay token. For the first time, the token's unfinished groove fills with a thin line of light.")
		s.updateQuests()
		s.persist()
		return
	}
	if s.Character.RoomID != "village_square" || !(strings.Contains(recipient, "bowl") || strings.Contains(recipient, "offering") || strings.Contains(recipient, "well")) {
		s.WriteLine("There is no suitable recipient here for that offering.")
		return
	}
	if !s.Character.HasStoryFlag("examined_scratched_coin") {
		s.WriteLine("The bowl waits beside the well. You should first understand the mark on the coin before deciding what it means.")
		return
	}
	if s.Character.HasStoryFlag("offered_coin_at_well") {
		s.WriteLine("The bowl is empty again, but the well's reflection has changed. It now shows a shoreline at dawn.")
		return
	}
	selected := -1
	for i, carried := range s.Character.Inventory {
		if strings.Contains(strings.ToLower(carried.Name), strings.ToLower(itemName)) && strings.Contains(strings.ToLower(carried.Name), "scratched bronze coin") {
			selected = i
			break
		}
	}
	if selected < 0 {
		s.WriteLine("You need to be carrying the scratched bronze coin to make this offering.")
		return
	}
	s.Character.Inventory = append(s.Character.Inventory[:selected], s.Character.Inventory[selected+1:]...)
	s.Character.SetStoryFlag("offered_coin_at_well")
	s.Character.Memories = append(s.Character.Memories, "I returned the scratched bronze coin to the offering bowl beside Asterion's well. The water showed a shore at dawn.")
	s.WriteLine("You lay the scratched coin in the clay bowl. It makes no sound against the bottom. In the well, black water draws back from a pale shore and a boat waiting beneath the first light.")
	s.WriteLine("For one breath, you hear an oar touch water. Then the reflection returns to the village square.")
	s.updateQuests()
	s.persist()
}

func (s *Session) offerMemory(args []string) {
	if s.Enemy != nil && s.Enemy.HP > 0 {
		s.WriteLine("The %s leaves you no room to offer a memory.", s.Enemy.Name)
		return
	}
	if s.Character.RoomID != "fields_asphodel" {
		s.WriteLine("There is no shade here to receive your memory.")
		return
	}
	var shadePresent bool
	for _, npc := range s.World.NPCs[s.Character.RoomID] {
		if npc.ID == "asphodel_shade" {
			shadePresent = true
			break
		}
	}
	if !shadePresent {
		s.WriteLine("The memory has no listener here.")
		return
	}
	if !s.Character.HasStoryFlag("charon_memory_heard") {
		s.WriteLine("You do not yet understand what the river asks of a memory. Speak with Charon at the Styx first.")
		return
	}
	if len(args) < 2 || !strings.EqualFold(args[0], "memory") {
		s.WriteLine("Offer a specific remembered experience, not just a name or a feeling.")
		s.WriteLine("Try: offer memory I remember someone singing off-key to comfort me when I feared the dark.")
		return
	}
	memory := strings.TrimSpace(strings.Join(args[1:], " "))
	if len(strings.Fields(memory)) < 6 {
		s.WriteLine("The shade listens, but the fragment is too small to hold. Describe a moment: who was there, what happened, and why it stayed with you.")
		return
	}
	if s.Character.HasStoryFlag("asphodel_memory_returned") {
		s.WriteLine("The shade touches the clay token. \"I will not ask you to prove it twice. The memory has a place here now.\"")
		return
	}
	s.Character.SetStoryFlag("asphodel_memory_returned")
	s.Character.Memories = append(s.Character.Memories, "I offered the Unremembered Shade this memory: "+memory)
	s.WriteLine("\x1b[1;36mThe shade listens without interrupting. When you finish, the pale flowers around its knees stir as if a breeze has crossed the field. The three grooves on the clay token darken: one straight, one broken, one unfinished.\x1b[0m")
	s.WriteLine("The shade says, \"I cannot tell whether this happened to me. Perhaps that is not the point. For the first time, I can imagine that an ordinary life leaves a mark even when its name is lost. I will keep the shape of what you told me.\"")
	s.WriteLine("\x1b[1;33mYour memory has become part of your soul's record. A Memory for the Dead is complete.\x1b[0m")
	s.updateQuests()
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

