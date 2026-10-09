package session

import (
	"io"
	"net"
	"strings"
	"testing"

	"fatewalker/game/character"
	"fatewalker/game/item"
	"fatewalker/game/combat"
	"fatewalker/game/quest"
	"fatewalker/world"
)

func newChoiceTestSession(t *testing.T) *Session {
	t.Helper()
	server, client := net.Pipe()
	t.Cleanup(func() {
		server.Close()
		client.Close()
	})
	go func() {
		_, _ = io.Copy(io.Discard, client)
	}()
	c := character.New("Testwalker")
	c.RoomID = "oracle_path"
	return &Session{Character: c, Conn: server, World: world.NewWorld()}
}

func TestOracleTrustChoiceLeavesPersistentOath(t *testing.T) {
	s := newChoiceTestSession(t)
	s.choose([]string{"trust"})

	if !s.Character.HasStoryFlag("oracle_choice_made") || !s.Character.HasStoryFlag("oracle_trust") {
		t.Fatal("trust choice flags were not recorded")
	}
	if got := s.Character.Reputation["delphi"]; got != 2 {
		t.Fatalf("Delphi reputation = %d, want 2", got)
	}
	if len(s.Character.Oaths) != 1 || len(s.Character.Memories) != 1 {
		t.Fatal("trust choice should leave an oath and a memory")
	}

	s.choose([]string{"defy"})
	if s.Character.HasStoryFlag("oracle_defied") {
		t.Fatal("a second choice should not overwrite the first")
	}
	if got := s.Character.Reputation["delphi"]; got != 2 {
		t.Fatalf("second choice changed reputation to %d, want 2", got)
	}
}

func TestOracleDefianceLeavesPersistentScar(t *testing.T) {
	s := newChoiceTestSession(t)
	s.choose([]string{"defy"})

	if !s.Character.HasStoryFlag("oracle_choice_made") || !s.Character.HasStoryFlag("oracle_defied") {
		t.Fatal("defiance choice flags were not recorded")
	}
	if got := s.Character.Reputation["delphi"]; got != -1 {
		t.Fatalf("Delphi reputation = %d, want -1", got)
	}
	if len(s.Character.Scars) != 1 || len(s.Character.Echoes) != 1 {
		t.Fatal("defiance should leave a scar and an echo")
	}
}

func TestDefeatEnemyAwardsQuestProgressForCombatKills(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_foothills"
	s.Enemy = &combat.Enemy{Name: "Harpy", HP: 0, XP: 70}
	s.defeatEnemy(false)

	if got := s.Character.Quests["black_thread"]; got != 1 {
		t.Fatalf("black_thread progress = %d, want 1", got)
	}
	if s.Character.Experience != 150 {
		t.Fatalf("experience = %d, want 150 (70 enemy XP + 80 quest XP)", s.Character.Experience)
	}
	if s.Enemy != nil {
		t.Fatal("defeated enemy should be cleared")
	}
}

func TestManticoreDropWorksWithStarterInventoryAndDoesNotDuplicate(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Level = 5
	s.Enemy = &combat.Enemy{Name: "Manticore", HP: 0, XP: 0}
	s.defeatEnemy(true)

	count := 0
	for _, owned := range s.Character.Inventory {
		if owned.Name == "manticore fang" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("manticore fang count = %d, want 1", count)
	}

	s.Enemy = &combat.Enemy{Name: "Manticore", HP: 0, XP: 0}
	s.defeatEnemy(true)
	count = 0
	for _, owned := range s.Character.Inventory {
		if owned.Name == "manticore fang" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("repeated Manticore kill duplicated fang; count = %d", count)
	}
}

func TestOracleChoiceUnlocksOnlyItsBranchQuest(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "temple_dawn"
	s.updateQuests()
	if got := s.Character.Quests["oath_across_the_river"]; got != 0 {
		t.Fatalf("locked oath quest progress = %d, want 0", got)
	}

	s.Character.SetStoryFlag("oracle_trust")
	s.updateQuests()
	if got := s.Character.Quests["oath_across_the_river"]; got != 1 {
		t.Fatalf("trusted branch quest progress = %d, want 1", got)
	}
	if got := s.Character.Quests["unwritten_path"]; got != 0 {
		t.Fatalf("defiance branch quest progress = %d, want 0", got)
	}
	if s.Character.Experience != 180 {
		t.Fatalf("branch quest XP = %d, want 180", s.Character.Experience)
	}
}

func TestDefianceUnlocksItsOwnQuest(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "ancient_athens"
	s.Character.SetStoryFlag("oracle_defied")
	s.updateQuests()
	if got := s.Character.Quests["unwritten_path"]; got != 1 {
		t.Fatalf("defiance quest progress = %d, want 1", got)
	}
	if got := s.Character.Quests["oath_across_the_river"]; got != 0 {
		t.Fatalf("trust branch quest progress = %d, want 0", got)
	}
}

func TestBlackThreadOnlyCountsHarpyKillsInFoothills(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "ancient_athens"
	s.advanceQuestKill("Harpy")
	if got := s.Character.Quests["black_thread"]; got != 0 {
		t.Fatalf("black_thread progress outside foothills = %d, want 0", got)
	}
	s.Character.RoomID = "olympus_foothills"
	s.advanceQuestKill("Harpy")
	if got := s.Character.Quests["black_thread"]; got != 1 {
		t.Fatalf("black_thread progress in foothills = %d, want 1", got)
	}
}

func TestEnteringFoothillsDoesNotCompleteBlackThread(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_foothills"
	s.updateQuests()
	if got := s.Character.Quests["black_thread"]; got != 0 {
		t.Fatalf("black_thread progress on arrival = %d, want 0 until a Harpy is defeated", got)
	}
	if s.Character.Experience != 0 {
		t.Fatalf("experience on arrival = %d, want 0", s.Character.Experience)
	}
}

func TestRoomQuestCompletesOnArrival(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "delphi_sanctum"
	s.updateQuests()
	if got := s.Character.Quests["oracle_whisper"]; got != 1 {
		t.Fatalf("oracle_whisper progress = %d, want 1", got)
	}
	if s.Character.Experience != 210 {
		t.Fatalf("experience = %d, want 210 (Oracle's Whisper plus Trials of the Seer)", s.Character.Experience)
	}
}


func TestQuestRewardsFactionReputation(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_foothills"
	s.advanceQuestKill("Harpy")
	if got := s.Character.Reputation["olympians"]; got != 1 {
		t.Fatalf("Olympian reputation after Black Thread = %d, want 1", got)
	}

	s.Character.SetStoryFlag("oracle_trust")
	s.Character.RoomID = "temple_dawn"
	s.updateQuests()
	if got := s.Character.Reputation["delphi"]; got != 1 {
		t.Fatalf("Delphi reputation after oath quest = %d, want 1", got)
	}
}

func TestFactionStandingChangesMerchantPrice(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_foothills"
	n := s.currentNPC(1)
	if got := s.merchantPrice(n, n.Shop[0]); got != 75 {
		t.Fatalf("neutral merchant price = %d, want 75", got)
	}
	s.Character.Reputation["olympians"] = 3
	if got := s.merchantPrice(n, n.Shop[0]); got != 67 {
		t.Fatalf("friendly merchant price = %d, want 67 (10%% discount rounded down)", got)
	}
	s.Character.Reputation["olympians"] = -2
	if got := s.merchantPrice(n, n.Shop[0]); got != 83 {
		t.Fatalf("hostile merchant price = %d, want 83 (10%% surcharge rounded up)", got)
	}
}

func TestDialogueMatchesLongerNaturalLanguageTopics(t *testing.T) {
	w := world.NewWorld()
	pythia := w.NPCs["oracle_path"][0]
	if got := pythia.DialogueFor("can you tell me about rebirth please"); got == "" {
		t.Fatal("longer natural-language topic should match rebirth dialogue")
	}
	if got := pythia.DialogueFor("WHAT DO YOU KNOW ABOUT THE THREAD?"); got == "" {
		t.Fatal("dialogue matching should ignore case and punctuation around keywords")
	}
}


func TestFactionStandingChangesNPCGreeting(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_foothills"
	n := s.currentNPC()
	s.Character.Reputation["olympians"] = 3
	if got := s.factionGreeting(n); got == "" {
		t.Fatal("friendly faction standing should change the merchant greeting")
	}
	s.Character.Reputation["olympians"] = -2
	if got := s.factionGreeting(n); got == "" {
		t.Fatal("hostile faction standing should change the merchant greeting")
	}
	s.Character.Reputation["olympians"] = 0
	if got := s.factionGreeting(n); got != "" {
		t.Fatalf("neutral faction greeting = %q, want empty to use default dialogue", got)
	}
}


func TestRiverQuestRequiresOracleProphecy(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "styx_shore"
	s.updateQuests()
	if got := s.Character.Quests["river_of_memory"]; got != 0 {
		t.Fatalf("river quest progress before oracle prophecy = %d, want 0", got)
	}
	s.Character.Quests["oracle_whisper"] = 1
	s.updateQuests()
	if got := s.Character.Quests["river_of_memory"]; got != 1 {
		t.Fatalf("river quest progress after oracle prophecy = %d, want 1", got)
	}
	if !s.Character.HasStoryFlag("styx_memory_recovered") {
		t.Fatal("completing river quest should record its persistent soul discovery")
	}
}

func TestRiverMemoryUnlocksModernEraQuest(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "styx_shore"
	s.Character.Quests["oracle_whisper"] = 1
	s.updateQuests()
	if !s.Character.HasStoryFlag("styx_memory_recovered") {
		t.Fatal("river memory was not recorded")
	}
	s.Character.RebirthTo("modern", "modern_crossroads")
	s.Character.RoomID = "modern_styx"
	s.updateQuests()
	if got := s.Character.Quests["echoes_in_glass"]; got != 1 {
		t.Fatalf("modern era quest progress = %d, want 1", got)
	}
	if !s.Character.HasStoryFlag("quest_completed_echoes_in_glass") {
		t.Fatal("modern quest completion should be recorded on the soul")
	}
}

func TestMyrtoRecognizesPersistentStyxMemory(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "modern_crossroads"
	s.Character.SetStoryFlag("styx_memory_recovered")
	n := s.currentNPC()
	if n == nil || n.ID != "athens_vendor" {
		t.Fatal("expected Myrto in modern crossroads")
	}
	greeting := s.persistentNPCGreeting(n, "hello")
	if greeting == "" || !strings.Contains(greeting, "river beneath the old world") {
		t.Fatalf("Myrto greeting = %q, want her to recognize the Styx memory", greeting)
	}
	if got := s.persistentNPCGreeting(n, "wares"); got != "" {
		t.Fatalf("non-greeting topic returned special dialogue %q", got)
	}
}

func TestMyrtoRecognizesPersistentOracleProphecy(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "modern_crossroads"
	s.Character.SetStoryFlag("quest_completed_oracle_whisper")
	n := s.currentNPC()
	greeting := s.persistentNPCGreeting(n, "greeting")
	if greeting == "" || !strings.Contains(greeting, "old prophecy") {
		t.Fatalf("Myrto greeting = %q, want her to recognize the Oracle prophecy", greeting)
	}
}


func TestDelphiTrialsGrantPersistentLaurelRelic(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "delphi_sanctum"
	s.updateQuests()
	if !s.Character.HasStoryFlag("delphi_trials_complete") {
		t.Fatal("Delphi trials should record a persistent story flag")
	}
	count := 0
	for _, owned := range s.Character.Inventory { if owned.Relic && owned.Name == "Laurel of the Seer" { count++ } }
	if count != 1 { t.Fatalf("Laurel of the Seer count = %d, want 1", count) }
	s.Character.RebirthTo("modern", "modern_crossroads")
	if !s.Character.HasStoryFlag("delphi_trials_complete") { t.Fatal("Delphi trial flag did not persist across rebirth") }
}

func TestSoulRelicSummaryReportsKnownEffect(t *testing.T) {
	inventory := []item.Item{
		{Name: "bronze sword", Kind: "sword"},
		{Name: "Styxglass Shard", Kind: "relic", Relic: true},
		{Name: "Another Relic", Kind: "relic", Relic: true},
	}
	if got := soulRelicCount(inventory); got != 2 {
		t.Fatalf("soulRelicCount = %d, want 2", got)
	}
	if got := soulRelicEffect("Styxglass Shard"); !strings.Contains(got, "8 incoming damage") {
		t.Fatalf("Styxglass effect = %q, want its defensive effect", got)
	}
	if got := soulRelicEffect("Unknown Relic"); !strings.Contains(got, "unknown") {
		t.Fatalf("unknown relic effect = %q, want an unknown-purpose message", got)
	}
}

func TestStyxglassWardAbsorbsDamageOncePerEncounter(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Inventory = append(s.Character.Inventory, item.Item{Name: "Styxglass Shard", Tier: item.Epic, Kind: "relic", Relic: true})

	remaining, absorbed := s.applySoulRelicWard(12)
	if remaining != 4 || absorbed != 8 {
		t.Fatalf("first ward = remaining %d, absorbed %d; want 4 and 8", remaining, absorbed)
	}
	remaining, absorbed = s.applySoulRelicWard(12)
	if remaining != 12 || absorbed != 0 {
		t.Fatalf("second ward = remaining %d, absorbed %d; want 12 and 0", remaining, absorbed)
	}
}

func TestStyxglassWardCannotAbsorbMoreThanIncomingDamage(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Inventory = append(s.Character.Inventory, item.Item{Name: "Styxglass Shard", Tier: item.Epic, Kind: "relic", Relic: true})

	remaining, absorbed := s.applySoulRelicWard(3)
	if remaining != 0 || absorbed != 3 {
		t.Fatalf("small-hit ward = remaining %d, absorbed %d; want 0 and 3", remaining, absorbed)
	}
}

func TestBranchQuestsGrantDistinctPersistentSoulRelics(t *testing.T) {
	s := newChoiceTestSession(t)
	s.completeQuest(quest.Quest{ID: "oath_across_the_river"})
	s.completeQuest(quest.Quest{ID: "oath_across_the_river"})
	s.completeQuest(quest.Quest{ID: "unwritten_path"})
	s.completeQuest(quest.Quest{ID: "unwritten_path"})
	counts := map[string]int{}
	for _, owned := range s.Character.Inventory {
		if owned.Relic { counts[owned.Name]++ }
	}
	if counts["Oracle's Thread"] != 1 || counts["Unwritten Ember"] != 1 {
		t.Fatalf("branch relic counts = %#v, want one of each branch relic", counts)
	}
	s.Character.RebirthTo("modern", "modern_crossroads")
	for _, name := range []string{"Oracle's Thread", "Unwritten Ember"} {
		found := false
		for _, owned := range s.Character.Inventory {
			if owned.Relic && owned.Name == name { found = true }
		}
		if !found { t.Fatalf("%s did not persist through rebirth", name) }
	}
}

func TestOracleThreadWardAndStyxglassChooseStrongestWard(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Inventory = append(s.Character.Inventory,
		item.Item{Name: "Oracle's Thread", Kind: "relic", Relic: true},
		item.Item{Name: "Styxglass Shard", Kind: "relic", Relic: true},
	)
	remaining, absorbed := s.applySoulRelicWard(12)
	if remaining != 4 || absorbed != 8 {
		t.Fatalf("strongest ward = remaining %d, absorbed %d; want 4 and 8", remaining, absorbed)
	}
	s.RelicWardSpent = false
	s.Character.Inventory = []item.Item{{Name: "Oracle's Thread", Kind: "relic", Relic: true}}
	remaining, absorbed = s.applySoulRelicWard(3)
	if remaining != 0 || absorbed != 3 {
		t.Fatalf("Oracle's Thread ward = remaining %d, absorbed %d; want 0 and 3", remaining, absorbed)
	}
}

func TestUnwrittenEmberAddsDamageOncePerEncounter(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Inventory = append(s.Character.Inventory, item.Item{Name: "Unwritten Ember", Kind: "relic", Relic: true})
	if damage, bonus := s.applySoulRelicStrike(0); damage != 0 || bonus != 0 || s.RelicStrikeSpent {
		t.Fatal("zero-damage attempt should not spend the Unwritten Ember")
	}
	damage, bonus := s.applySoulRelicStrike(14)
	if damage != 20 || bonus != 6 || !s.RelicStrikeSpent {
		t.Fatalf("first strike = damage %d, bonus %d, spent %v; want 20, 6, true", damage, bonus, s.RelicStrikeSpent)
	}
	damage, bonus = s.applySoulRelicStrike(14)
	if damage != 14 || bonus != 0 {
		t.Fatalf("second strike = damage %d, bonus %d; want 14 and 0", damage, bonus)
	}
}

func TestStyxglassSoulRelicPersistsAcrossRebirthWithoutDuplication(t *testing.T) {
	s := newChoiceTestSession(t)
	reward := quest.Quest{ID: "river_of_memory", RewardXP: 1}
	s.completeQuest(reward)
	s.completeQuest(reward)

	count := 0
	for _, owned := range s.Character.Inventory {
		if owned.Name == "Styxglass Shard" && owned.Relic {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("Styxglass Shard count after repeated reward = %d, want 1", count)
	}

	s.Character.RebirthTo("modern", "modern_crossroads")
	count = 0
	for _, owned := range s.Character.Inventory {
		if owned.Name == "Styxglass Shard" && owned.Relic {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("Styxglass Shard count after rebirth = %d, want 1", count)
	}
}

func TestLifeGiftCanOnlyBeUsedOncePerLife(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.LifeGift = "Echo Sight"
	s.Character.Mana = 0
	s.invokeGift()
	if s.Character.Mana != 12 {
		t.Fatalf("Echo Sight mana = %d, want 12", s.Character.Mana)
	}
	if !s.Character.LifeGiftUsed {
		t.Fatal("using life-gift should mark it spent")
	}
	s.Character.Mana = 0
	s.invokeGift()
	if s.Character.Mana != 0 {
		t.Fatalf("second Echo Sight use restored mana to %d, want 0", s.Character.Mana)
	}
	s.Character.RebirthTo("modern", "modern_crossroads")
	if s.Character.LifeGiftUsed {
		t.Fatal("life-gift should refresh after rebirth")
	}
}


func TestAwardExperienceLevelsCharacterAndRestoresResources(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Experience = 55
	s.Character.HP = 20
	s.Character.Mana = 2
	s.awardExperience(10)
	if s.Character.Level != 2 {
		t.Fatalf("level after XP award = %d, want 2", s.Character.Level)
	}
	if s.Character.HP != s.Character.MaxHP || s.Character.Mana != s.Character.MaxMana {
		t.Fatalf("level-up resources = HP %d/%d Mana %d/%d, want fully restored", s.Character.HP, s.Character.MaxHP, s.Character.Mana, s.Character.MaxMana)
	}
}

func TestStrongestOffensiveSoulRelicWinsRegardlessOfInventoryOrder(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Inventory = append(s.Character.Inventory,
		item.Item{Name: "Laurel of the Seer", Kind: "relic", Relic: true},
		item.Item{Name: "Unwritten Ember", Kind: "relic", Relic: true},
	)
	damage, bonus := s.applySoulRelicStrike(14)
	if damage != 20 || bonus != 6 {
		t.Fatalf("strongest relic strike = damage %d, bonus %d; want 20 and 6", damage, bonus)
	}
	if got := s.soulRelicStrikeName(); got != "Unwritten Ember" {
		t.Fatalf("announced relic = %q, want Unwritten Ember", got)
	}
}


func TestMyrtoGreetingReflectsOracleTrust(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.SetStoryFlag("oracle_trust")
	s.Character.RebirthTo("modern", "modern_crossroads")
	got := s.persistentNPCGreeting(s.currentNPC(), "hello")
	if !strings.Contains(got, "trusted you with a warning") {
		t.Fatalf("unexpected modern greeting: %q", got)
	}
}


func TestMyrtoGreetingReflectsOracleDefiance(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.SetStoryFlag("oracle_defied")
	s.Character.RebirthTo("modern", "modern_crossroads")
	got := s.persistentNPCGreeting(s.currentNPC(), "hello")
	if !strings.Contains(got, "mistake prophecy for permission") {
		t.Fatalf("unexpected modern greeting: %q", got)
	}
}


func TestLevelTenRebirthOpensModernAthensAndPreservesSoul(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.SetStoryFlag("oracle_trust")
	s.Character.Inventory = append(s.Character.Inventory, item.Item{Name: "Styxglass Shard", Kind: "relic", Relic: true})
	s.awardExperience(3645)
	if s.Character.Level != 10 {
		t.Fatalf("level after reaching first rebirth threshold = %d, want 10", s.Character.Level)
	}
	s.rebirth(nil)
	if s.Character.Life != 2 || s.Character.Era != "modern" || s.Character.RoomID != "modern_crossroads" {
		t.Fatalf("rebirth destination = life %d, era %q, room %q", s.Character.Life, s.Character.Era, s.Character.RoomID)
	}
	if s.Character.Level != 1 || s.Character.Experience != 0 {
		t.Fatalf("new life progression = level %d, XP %d; want level 1 and 0 XP", s.Character.Level, s.Character.Experience)
	}
	if !s.Character.HasStoryFlag("oracle_trust") {
		t.Fatal("Oracle choice did not survive rebirth")
	}
	foundRelic := false
	for _, owned := range s.Character.Inventory {
		if owned.Relic && owned.Name == "Styxglass Shard" {
			foundRelic = true
		}
	}
	if !foundRelic {
		t.Fatal("soul relic did not survive rebirth")
	}
	if s.Character.LifeGift != "Echo Sight" || s.Character.LifeGiftUsed {
		t.Fatalf("new life gift = %q, used=%v; want unused Echo Sight", s.Character.LifeGift, s.Character.LifeGiftUsed)
	}
}


func TestStyxRecoveryRestoresCharacterAtGates(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.HP = 1
	s.Enemy = &combat.Enemy{Name: "Test Colossus", Level: 99, Damage: 1000, HP: 500, MaxHP: 500}
	s.enemyTurn()
	if s.Character.HP != s.Character.MaxHP || s.Character.Mana != s.Character.MaxMana { t.Fatal("recovery did not restore resources") }
	if s.Character.RoomID != "olympus_gates" || s.Enemy != nil { t.Fatalf("recovery state = room %q, enemy %#v", s.Character.RoomID, s.Enemy) }
	if s.Character.Life != 1 { t.Fatalf("life = %d, want 1", s.Character.Life) }
	if s.RelicWardSpent || s.RelicStrikeSpent { t.Fatal("relic encounter state was not reset") }
}


func TestRoomDiscoveryRewardIsOneTime(t *testing.T) {
	s := newChoiceTestSession(t)
	s.recordDiscovery("delphi_sanctum")
	first := s.Character.Experience
	s.recordDiscovery("delphi_sanctum")
	if first != 110 || s.Character.Experience != first { t.Fatalf("XP after discovery/revisit = %d/%d, want 110/110", first, s.Character.Experience) }
	if !s.Character.HasStoryFlag("room_discovered_delphi_sanctum") { t.Fatal("discovery flag missing") }
}


func TestJournalSelectsNextUnlockedQuest(t *testing.T) {
	s := newChoiceTestSession(t)
	q, ok := s.nextAvailableQuest()
	if !ok || q.ID != "black_thread" { t.Fatalf("first suggested quest = %q, found=%v", q.ID, ok) }
	s.Character.Quests["black_thread"] = 1
	q, ok = s.nextAvailableQuest()
	if !ok || q.ID != "oracle_whisper" { t.Fatalf("next suggested quest = %q, found=%v", q.ID, ok) }
}


func TestRestRestoresResourcesOnlyOutsideCombat(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.HP = 25
	s.Character.Mana = 3
	s.Enemy = &combat.Enemy{Name: "Harpy", HP: 20, MaxHP: 20}
	s.rest()
	if s.Character.HP != 25 || s.Character.Mana != 3 { t.Fatal("rest should not work during combat") }
	s.Enemy = nil
	s.rest()
	if s.Character.HP != s.Character.MaxHP || s.Character.Mana != s.Character.MaxMana { t.Fatal("rest did not restore health and mana") }
}


func TestCombatVictoryAwardsDrachmae(t *testing.T) {
	s := newChoiceTestSession(t)
	before := s.Character.Gold
	s.Enemy = &combat.Enemy{Name: "Harpy", Level: 2, HP: 0, XP: 0}
	s.defeatEnemy(false)
	if got := s.Character.Gold - before; got != 15 { t.Fatalf("gold reward = %d, want 15", got) }
}


func TestFutureAreasOpenAfterSecondRebirth(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Life = 2
	if s.canEnter("future_skyway") { t.Fatal("future path opened too early") }
	s.Character.Life = 3
	if !s.canEnter("future_skyway") { t.Fatal("future path stayed closed after rebirth") }
	if !s.canEnter("far_era") { t.Fatal("far era stayed closed after rebirth") }
}

func TestModernAndFutureRoomsSpawnEnemies(t *testing.T) {
	s := newChoiceTestSession(t)
	for _, tc := range []struct{ life int; room, name string }{
		{2, "modern_styx", "Styx Wraith"},
		{2, "modern_metro", "Echo Hound"},
		{3, "future_city", "Chronal Warden"},
		{3, "future_moon", "Moonshade"},
		{3, "far_era", "Last Shore Titan"},
	} {
		s.Character.Life = tc.life
		s.Character.RoomID = tc.room
		s.Enemy = nil
		s.spawnEnemy()
		if s.Enemy == nil || s.Enemy.Name != tc.name || s.Enemy.XP <= 0 { t.Errorf("unexpected enemy for %s: %#v", tc.room, s.Enemy) }
	}
}


func TestCerberusAndTartarusSpawnInTheirRooms(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Level = 9
	for _, tc := range []struct{ room, name string }{{"cerberus_gate", "Cerberus"}, {"tartarus_edge", "Tartarus Brute"}} {
		s.Character.RoomID = tc.room
		s.Enemy = nil
		s.spawnEnemy()
		if s.Enemy == nil || s.Enemy.Name != tc.name { t.Errorf("room %s spawned %#v", tc.room, s.Enemy) }
	}
}


func TestMovementCommandsSupportHiddenEntrances(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Life = 2
	s.Character.RoomID = "modern_acropolis"
	s.handleCommand("in")
	if s.Character.RoomID != "modern_sanctum" { t.Fatalf("in command moved to %q", s.Character.RoomID) }
	s.Enemy = nil
	s.handleCommand("out")
	if s.Character.RoomID != "modern_acropolis" { t.Fatalf("out command moved to %q", s.Character.RoomID) }
}


func TestCerberusQuestRequiresStyxMemoryAndRewardsFavor(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "cerberus_gate"
	s.advanceQuestKill("Cerberus")
	if s.Character.HasStoryFlag("cerberus_defeated") { t.Fatal("Cerberus quest should be locked before Styx memory") }
	s.Character.SetStoryFlag("styx_memory_recovered")
	s.advanceQuestKill("Cerberus")
	if !s.Character.HasStoryFlag("cerberus_defeated") { t.Fatal("Cerberus victory was not recorded") }
	found := false
	for _, favor := range s.Character.Favors { if favor == "Cerberus Oath" { found = true } }
	if !found { t.Fatal("Cerberus reward favor missing") }
}


func TestOracleChoiceWaitsUntilCombatEnds(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Enemy = &combat.Enemy{Name: "Satyr", HP: 20, MaxHP: 20}
	s.choose([]string{"trust"})
	if s.Character.HasStoryFlag("oracle_choice_made") { t.Fatal("choice was accepted during combat") }
	s.Enemy = nil
	s.choose([]string{"trust"})
	if !s.Character.HasStoryFlag("oracle_trust") { t.Fatal("choice was not accepted after combat") }
}


func TestFirstLifeRouteCompletesCoreQuests(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_gates"
	s.Character.Level = 10
	s.Character.Experience = 3645
	s.move("north")
	s.Enemy.HP = 0
	s.defeatEnemy(false)
	s.move("north")
	s.Enemy.HP = 0
	s.defeatEnemy(false)
	s.choose([]string{"trust"})
	s.move("north")
	if !s.Character.HasStoryFlag("quest_completed_oracle_whisper") || !s.Character.HasStoryFlag("delphi_trials_complete") { t.Fatal("Delphi quests missing") }
	s.move("east")
	if !s.Character.HasStoryFlag("quest_completed_oath_across_the_river") { t.Fatal("trust quest missing") }
	s.move("down")
	if !s.Character.HasStoryFlag("styx_memory_recovered") { t.Fatal("Styx quest missing") }
	s.rebirth(nil)
	if s.Character.Life != 2 || s.Character.RoomID != "modern_crossroads" { t.Fatalf("rebirth destination: life %d room %q", s.Character.Life, s.Character.RoomID) }
}


func TestModernLifeUnlocksFutureLifeProgression(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Life = 2
	s.Character.Era = "modern"
	s.Character.RoomID = "modern_crossroads"
	s.Character.Level = 10
	s.Character.Experience = 3645
	s.Character.SetStoryFlag("styx_memory_recovered")

	s.move("north")
	s.Enemy.HP = 0
	s.defeatEnemy(false)
	s.move("east")
	s.Enemy.HP = 0
	s.defeatEnemy(false)
	s.move("east")
	if !s.Character.HasStoryFlag("quest_completed_echoes_in_glass") {
		t.Fatal("modern Styx quest did not complete")
	}
	s.Enemy.HP = 0
	s.defeatEnemy(false)

	s.rebirth([]string{"future"})
	if s.Character.Life != 3 || s.Character.Era != "future" || s.Character.RoomID != "future_city" {
		t.Fatalf("future rebirth = life %d era %q room %q", s.Character.Life, s.Character.Era, s.Character.RoomID)
	}
	s.look()
	if s.Enemy == nil || s.Enemy.Name != "Chronal Warden" {
		t.Fatalf("future starting encounter = %#v", s.Enemy)
	}
}


func TestRebirthWaitsForCombatToEnd(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Level = 10
	s.Enemy = &combat.Enemy{Name: "Harpy", HP: 10, MaxHP: 10}
	s.rebirth(nil)
	if s.Character.Life != 1 { t.Fatal("rebirth occurred during combat") }
	s.Enemy = nil
	s.rebirth(nil)
	if s.Character.Life != 2 { t.Fatal("rebirth did not work after combat") }
}


func TestMuseumQuestPersistsItsRewards(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Life = 2
	s.Character.RoomID = "modern_museum"
	s.Character.SetStoryFlag("styx_memory_recovered")
	s.updateQuests()
	if !s.Character.HasStoryFlag("museum_echo_uncovered") { t.Fatal("museum flag missing") }
	if len(s.Character.Memories) == 0 { t.Fatal("memory reward missing") }
}


func TestMyrtoAnswersAboutMuseumEcho(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "modern_crossroads"
	s.Character.SetStoryFlag("museum_echo_uncovered")
	got := s.persistentNPCGreeting(s.currentNPC(), "museum")
	if !strings.Contains(got, "behind glass") { t.Fatalf("museum response = %q", got) }
}


// TestNewSessionEnablesAutomaticMap verifies the default for fresh connections.
func TestNewSessionEnablesAutomaticMap(t *testing.T) {
	server, client := net.Pipe()
	t.Cleanup(func() { server.Close(); client.Close() })
	go func() { _, _ = io.Copy(io.Discard, client) }()

	s := New(nil, server, world.NewWorld())
	if !s.AutoMap {
		t.Fatal("automatic map should be enabled for new sessions")
	}
}

// TestAutomaticMapCommandTogglesState verifies the automap command and aliases.
func TestAutomaticMapCommandTogglesState(t *testing.T) {
	s := newChoiceTestSession(t)
	if s.AutoMap {
		t.Fatal("test helper should start with the zero-value automap setting")
	}

	s.handleCommand("automap on")
	if !s.AutoMap { t.Fatal("automap on did not enable automatic mapping") }
	s.handleCommand("automap off")
	if s.AutoMap { t.Fatal("automap off did not disable automatic mapping") }
	s.handleCommand("automap toggle")
	if !s.AutoMap { t.Fatal("automap toggle did not enable automatic mapping") }
	s.handleCommand("maptoggle")
	if s.AutoMap { t.Fatal("maptoggle alias did not toggle automatic mapping off") }
}
