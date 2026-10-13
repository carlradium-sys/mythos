package session

import (
	"fatewalker/game/power"
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

func TestStoryIntroIsRecordedForFirstLife(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "village_square"
	s.TutorialStep = 0
	s.storyIntro()
	if !s.Character.HasStoryFlag("tutorial_intro_seen") {
		t.Fatal("first-life story prologue should be recorded")
	}
	s.storyIntro()
	if !s.Character.HasStoryFlag("tutorial_intro_seen") {
		t.Fatal("story prologue flag should remain recorded")
	}
}

func TestVillageTutorialRequiresHintBeforeRoadTopic(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "village_lane"
	s.TutorialStep = 6
	s.talk([]string{"1", "road"})
	if s.TutorialStep != 6 {
		t.Fatal("road topic should not bypass the hint lesson")
	}
	s.hint(nil)
	if !s.Character.HasStoryFlag("tutorial_hint_used") {
		t.Fatal("asking for an NPC hint should record tutorial hint use")
	}
	s.talk([]string{"1", "road"})
	if s.TutorialStep != 7 {
		t.Fatal("road topic after asking for a hint should unlock the exits lesson")
	}
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
	s.Character.SetStoryFlag("room_discovered_delphi_sanctum")
	s.updateQuests()
	if got := s.Character.Quests["oracle_whisper"]; got != 1 {
		t.Fatalf("oracle_whisper progress = %d, want 1", got)
	}
	if s.Character.Experience != 120 {
		t.Fatalf("experience = %d, want 120 (Oracle's Whisper only; the inscription still needs examining)", s.Character.Experience)
	}
	if got := s.Character.Quests["delphi_trials"]; got != 0 {
		t.Fatalf("Delphi trials progress on room entry = %d, want 0 until the inscription is examined", got)
	}
	s.examineTarget([]string{"inscription"})
	if got := s.Character.Quests["delphi_trials"]; got != 1 {
		t.Fatalf("Delphi trials progress after examining inscription = %d, want 1", got)
	}
	if s.Character.Experience != 210 {
		t.Fatalf("experience after examining inscription = %d, want 210", s.Character.Experience)
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
	s.Character.RoomID = "therons_forge"
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
	s.Character.RoomID = "therons_forge"
	n := s.currentNPC(1)
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
	n := s.currentNPC(1)
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
	n := s.currentNPC(1)
	greeting := s.persistentNPCGreeting(n, "greeting")
	if greeting == "" || !strings.Contains(greeting, "old prophecy") {
		t.Fatalf("Myrto greeting = %q, want her to recognize the Oracle prophecy", greeting)
	}
}


func TestDelphiTrialsGrantPersistentLaurelRelic(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "delphi_sanctum"
	s.Character.SetStoryFlag("room_discovered_delphi_sanctum")
	s.updateQuests()
	s.examineTarget([]string{"inscription"})
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
	got := s.persistentNPCGreeting(s.currentNPC(1), "hello")
	if !strings.Contains(got, "trusted you with a warning") {
		t.Fatalf("unexpected modern greeting: %q", got)
	}
}


func TestMyrtoGreetingReflectsOracleDefiance(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.SetStoryFlag("oracle_defied")
	s.Character.RebirthTo("modern", "modern_crossroads")
	got := s.persistentNPCGreeting(s.currentNPC(1), "hello")
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
	s.EnemyAggressive = true
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


func TestNewCharacterBeginsInAsterionVillage(t *testing.T) {
	c := character.New("Newwalker")
	if c.RoomID != "village_square" {
		t.Fatalf("new character room = %q, want village_square", c.RoomID)
	}
}

func TestVillageTutorialRouteAndGuideExist(t *testing.T) {
	w := world.NewWorld()
	square := w.GetRoom("village_square")
	lane := w.GetRoom("village_lane")
	if square == nil || lane == nil {
		t.Fatal("village tutorial rooms are missing")
	}
	if square.Exits["north"] != lane.ID || lane.Exits["south"] != square.ID || lane.Exits["north"] != "olympus_foothills" {
		t.Fatal("village tutorial route is not connected correctly")
	}
	if len(w.NPCs["village_lane"]) == 0 || w.NPCs["village_lane"][0].ID != "village_guide" {
		t.Fatal("village lane must contain the tutorial guide")
	}
}

func TestVillageTutorialBlocksMountainRoadUntilGuideLessonCompletes(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "village_lane"
	s.TutorialStep = 0
	s.move("north")
	if s.Character.RoomID != "village_lane" {
		t.Fatalf("tutorial bypass moved to %q before lessons completed", s.Character.RoomID)
	}
	s.TutorialStep = 9
	s.move("north")
	if s.Character.RoomID != "olympus_foothills" {
		t.Fatalf("completed tutorial moved to %q, want olympus_foothills", s.Character.RoomID)
	}
}

func TestTutorialProgressesOneActionAtATime(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "village_lane"
	s.TutorialStep = 2
	s.lookTarget([]string{"Damon"})
	if !s.Character.HasStoryFlag("tutorial_inspected_npc") { t.Fatal("inspecting Damon should record the inspection lesson") }
	s.advanceTutorial(3)
	s.talk([]string{"1", "hello"})
	if s.TutorialStep != 4 { t.Fatal("greeting should advance to the village topic lesson") }
	s.talk([]string{"1", "village"})
	if s.TutorialStep != 5 { t.Fatal("village topic should advance to the hint lesson") }
	s.handleCommand("hint")
	if s.TutorialStep != 6 || !s.Character.HasStoryFlag("tutorial_hint_used") { t.Fatal("hint should advance to the road topic lesson") }
	s.talk([]string{"1", "road"})
	if s.TutorialStep != 7 { t.Fatal("road topic should advance to the exits lesson") }
	s.handleCommand("exits")
	if s.TutorialStep != 8 { t.Fatal("exits should advance to the map lesson") }
	s.handleCommand("map")
	if s.TutorialStep != 9 { t.Fatal("map should advance to the foothills lesson") }
	s.Character.RoomID = "olympus_foothills"
	s.TutorialStep = 10
	s.Enemy = &combat.Enemy{Name: "Harpy", Description: "A test harpy.", Level: 1, HP: 20, MaxHP: 20, XP: 0}
	s.handleCommand("look Harpy")
	if s.TutorialStep != 11 || !s.Character.HasStoryFlag("tutorial_enemy_inspected") { t.Fatal("inspecting the Harpy should advance to the fight lesson") }
	s.Enemy.HP = 0
	s.defeatEnemy(false)
	if s.TutorialStep != 12 || !s.Character.HasStoryFlag("tutorial_combat_complete") { t.Fatal("Harpy victory should advance to inventory") }
	s.handleCommand("inv")
	if s.TutorialStep != 13 { t.Fatal("inventory should advance to equipment") }
	s.handleCommand("eq")
	if s.TutorialStep != 14 { t.Fatal("equipment should advance to quests") }
	s.handleCommand("quests")
	if s.TutorialStep != 15 { t.Fatal("quests should advance to journal") }
	s.handleCommand("journal")
	if s.TutorialStep != 16 || !s.Character.HasStoryFlag("tutorial_complete") { t.Fatal("journal should complete the tutorial") }
}

func TestTutorialRequiresInspectionBeforeAttackingHarpy(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_foothills"
	s.TutorialStep = 11
	s.Enemy = &combat.Enemy{Name: "Harpy", Level: 1, HP: 30, MaxHP: 30, Damage: 1, XP: 0}
	s.attack(nil)
	if s.Enemy.HP != 30 {
		t.Fatal("tutorial should require inspecting the Harpy before the first attack")
	}
	s.lookTarget([]string{"Harpy"})
	if !s.Character.HasStoryFlag("tutorial_enemy_inspected") {
		t.Fatal("look Harpy did not record inspection")
	}
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
	s.examineTarget([]string{"inscription"})
	if !s.Character.HasStoryFlag("quest_completed_oracle_whisper") || !s.Character.HasStoryFlag("delphi_trials_complete") { t.Fatal("Delphi quests missing after examining the inscription") }
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
	s.examineTarget([]string{"mask"})
	if !s.Character.HasStoryFlag("museum_echo_uncovered") { t.Fatal("museum flag missing after examining the mask") }
	if len(s.Character.Memories) == 0 { t.Fatal("memory reward missing") }
}


func TestMyrtoAnswersAboutMuseumEcho(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "modern_crossroads"
	s.Character.SetStoryFlag("museum_echo_uncovered")
	for _, question := range []string{"museum", "can you tell me about the museum?", "tell me about artifacts"} {
		got := s.persistentNPCGreeting(s.currentNPC(1), question)
		if !strings.Contains(got, "behind glass") { t.Errorf("museum question %q response = %q", question, got) }
	}
}


// TestNewSessionKeepsAutomaticMapOffByDefault verifies the compact screen default.
func TestNewSessionKeepsAutomaticMapOffByDefault(t *testing.T) {
	server, client := net.Pipe()
	t.Cleanup(func() { server.Close(); client.Close() })
	go func() { _, _ = io.Copy(io.Discard, client) }()

	s := New(nil, server, world.NewWorld())
	if s.AutoMap {
		t.Fatal("automatic map should be off by default to keep the screen compact")
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


// Natural-language dialogue should resolve to the same intent as short topics.
func TestNPCDialogueMatchesNaturalLanguageQuestions(t *testing.T) {
	w := world.NewWorld()
	damon := w.NPCs["village_lane"][0]
	for _, phrase := range []string{
		"village",
		"tell me more about the village",
		"can you please tell me about the village",
		"where is the village?",
		"could you tell me about home",
	} {
		keyword, response := damon.MatchDialogue(phrase)
		if keyword != "village" && keyword != "home" {
			t.Errorf("phrase %q resolved to keyword %q", phrase, keyword)
		}
		if response == "" { t.Errorf("phrase %q did not resolve to a response", phrase) }
	}
}

func TestTalkByFirstNameAndNaturalQuestionAdvancesTutorial(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "village_lane"
	s.TutorialStep = 4
	s.handleCommand("talk Damon, can you please tell me about the village?")
	if s.TutorialStep != 5 {
		t.Fatalf("natural-language village question left tutorial at step %d", s.TutorialStep)
	}
	s.Character.SetStoryFlag("tutorial_hint_used")
	s.TutorialStep = 6
	s.handleCommand("talk Damon, where is the village road?")
	if s.TutorialStep != 7 {
		t.Fatalf("natural-language road question left tutorial at step %d", s.TutorialStep)
	}
}


func TestTutorialPracticeEnemiesArePassiveAndPersistentlyDefeated(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.SetStoryFlag("tutorial_available")
	s.Character.RoomID = "tutorial_yard"
	s.PrologueStep = 11
	s.look()
	if s.Enemy == nil || s.Enemy.Name != "Training Shade" {
		t.Fatal("tutorial courtyard should spawn the training shade")
	}
	if s.EnemyAggressive {
		t.Fatal("tutorial creature should not be aggressive before the player attacks")
	}
	hp := s.Character.HP
	s.enemyTurn()
	if s.Character.HP != hp {
		t.Fatal("passive tutorial creature attacked without being attacked")
	}
	s.Enemy.HP = 0
	s.defeatEnemy(false)
	if !s.Character.HasStoryFlag("tutorial_enemy_slain_tutorial_yard") {
		t.Fatal("defeated tutorial creature should be persisted as slain")
	}
	s.look()
	if s.Enemy != nil {
		t.Fatal("slain tutorial creature respawned after looking at the room")
	}
}

func TestTutorialProgressIsSavedAsCharacterFlag(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.SetStoryFlag("tutorial_available")
	s.Character.RoomID = "tutorial_shore"
	s.PrologueStep = 0
	s.advancePrologueForCommand([]string{"look"})
	if s.PrologueStep != 1 || !s.Character.HasStoryFlag("tutorial_lesson_1") {
		t.Fatal("completing a tutorial lesson should persist the next lesson flag")
	}
}

func TestTutorialRoomsAreLockedForCompletedCharacters(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.SetStoryFlag("tutorial_complete")
	if s.canEnter("tutorial_shore") {
		t.Fatal("completed characters must not be allowed back into tutorial rooms")
	}
}


func TestConsumableRestoresHealthAndIsConsumed(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "village_square"
	s.Character.HP = 40
	s.Character.Inventory = append(s.Character.Inventory, item.Item{
		Name: "laurel poultice", Kind: "consumable", Healing: 20,
	})
	s.useItem([]string{"laurel", "poultice"})
	if s.Character.HP != 60 {
		t.Fatalf("health after poultice = %d, want 60", s.Character.HP)
	}
	for _, it := range s.Character.Inventory {
		if it.Name == "laurel poultice" {
			t.Fatal("consumable remained in inventory after use")
		}
	}
}

func TestConsumableDoesNotOverhealOrWasteItemAtFullHealth(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.HP = s.Character.MaxHP - 5
	s.Character.Inventory = append(s.Character.Inventory, item.Item{
		Name: "ambrosia draught", Kind: "consumable", Healing: 45,
	})
	s.useItem([]string{"ambrosia", "draught"})
	if s.Character.HP != s.Character.MaxHP {
		t.Fatalf("health = %d, want max %d", s.Character.HP, s.Character.MaxHP)
	}
	if len(s.Character.Inventory) != 2 {
		t.Fatalf("inventory length = %d, want 2 after effective use", len(s.Character.Inventory))
	}

	s.Character.Inventory = append(s.Character.Inventory, item.Item{
		Name: "second ambrosia draught", Kind: "consumable", Healing: 45,
	})
	s.useItem([]string{"second", "ambrosia"})
	if len(s.Character.Inventory) != 3 {
		t.Fatal("an ineffective consumable should not be consumed at full health")
	}
}

func TestConsumableUseCostsEnemyTurnDuringAggressiveCombat(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.HP = 20
	s.Character.Inventory = append(s.Character.Inventory, item.Item{
		Name: "ambrosia draught", Kind: "consumable", Healing: 45,
	})
	s.Enemy = &combat.Enemy{Name: "Practice Foe", Level: 20, HP: 50, MaxHP: 50, Damage: 10}
	s.EnemyAggressive = true
	s.useItem([]string{"ambrosia", "draught"})
	if len(s.Character.Inventory) != 2 {
		t.Fatal("used consumable should be removed from inventory")
	}
	if s.Character.HP >= 65 {
		t.Fatalf("enemy did not take its turn after the item was used; HP = %d", s.Character.HP)
	}
}


func TestCraftingConsumesMaterialAndGoldAtTheForge(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "therons_forge"
	s.Character.SetStoryFlag("forge_discovered")
	s.Character.Gold = 100
	s.Character.Inventory = append(s.Character.Inventory, item.Item{
		Name: "harpy feather", Kind: "material",
	})
	s.craft([]string{"stormguard", "cuirass"})
	if s.Character.Gold != 57 {
		t.Fatalf("gold after crafting and quest reward = %d, want 57", s.Character.Gold)
	}
	foundArmor, foundMaterial := false, false
	for _, it := range s.Character.Inventory {
		if it.Name == "stormguard cuirass" && it.Armor == 7 { foundArmor = true }
		if it.Name == "harpy feather" { foundMaterial = true }
	}
	if !foundArmor { t.Fatal("crafted armor was not added to inventory") }
	if foundMaterial { t.Fatal("crafting ingredient was not consumed") }
	if s.Character.Quests["forge_remembers"] != 1 { t.Fatal("crafting should complete The Forge Remembers quest") }
}

func TestCraftingRequiresForgeAndEnoughGold(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "village_square"
	s.Character.Gold = 100
	s.Character.Inventory = append(s.Character.Inventory, item.Item{Name:"harpy feather", Kind:"material"})
	s.craft([]string{"stormguard", "cuirass"})
	if s.Character.Gold != 100 || len(s.Character.Inventory) != 3 {
		t.Fatal("crafting outside the forge should not consume gold or materials")
	}
	s.Character.RoomID = "therons_forge"
	s.Character.Gold = 10
	s.craft([]string{"stormguard", "cuirass"})
	if s.Character.Gold != 10 || len(s.Character.Inventory) != 3 {
		t.Fatal("crafting without sufficient gold should not consume ingredients")
	}
}


func TestGuardReducesIncomingDamageAndIsConsumed(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.HP = 100
	s.Enemy = &combat.Enemy{Name: "Test Foe", Level: 1, HP: 50, MaxHP: 50, Damage: 20}
	s.EnemyAggressive = true
	s.guard()
	if s.Guarding {
		t.Fatal("guard should be consumed by the enemy's next turn")
	}
	if s.Character.HP < 90 || s.Character.HP > 100 {
		t.Fatalf("guarded health = %d; expected at most 10 damage", s.Character.HP)
	}
}


// TestTutorialSatyrStaysDefeatedAndAllowsThePlayerToLeave verifies the
// tutorial encounter cannot respawn after victory and block the south exit.
func TestTutorialSatyrStaysDefeatedAndAllowsThePlayerToLeave(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "tutorial_grove"
	s.Character.SetStoryFlag("tutorial_available")
	s.PrologueStep = 20
	s.Enemy = &combat.Enemy{
		Name: "Practice Satyr", HP: 0, MaxHP: 28, Level: 1, Damage: 3, XP: 0,
	}

	s.defeatEnemy(false)
	if !s.Character.HasStoryFlag("tutorial_enemy_slain_tutorial_grove") {
		t.Fatal("defeating the practice satyr should record the tutorial victory")
	}

	s.advancePrologueForCommand([]string{"attack", "satyr"})
	if s.PrologueStep != 21 {
		t.Fatalf("prologue step after satyr victory = %d, want 21", s.PrologueStep)
	}
	s.look()
	if s.Enemy != nil {
		t.Fatalf("practice satyr respawned after victory: %#v", s.Enemy)
	}

	s.advancePrologueForCommand([]string{"look"})
	if s.PrologueStep != 22 {
		t.Fatalf("prologue step after looking at cleared grove = %d, want 22", s.PrologueStep)
	}
	s.move("south")
	if s.Character.RoomID != "tutorial_yard" {
		t.Fatalf("room after leaving cleared grove = %q, want tutorial_yard", s.Character.RoomID)
	}
}


func TestFleeDoesNotImmediatelyRespawnEnemy(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "oracle_path"
	s.Character.HP = s.Character.MaxHP
	s.Enemy = &combat.Enemy{Name: "Satyr", Level: 1, HP: 40, MaxHP: 40, Damage: 1}
	s.EnemyAggressive = true

	s.flee()
	if s.Enemy != nil {
		t.Fatalf("enemy should be cleared after a successful escape, got %#v", s.Enemy)
	}
	if s.EncounterClearedRoom != "oracle_path" {
		t.Fatalf("cleared encounter room = %q, want oracle_path", s.EncounterClearedRoom)
	}
	s.look()
	if s.Enemy != nil {
		t.Fatalf("enemy respawned immediately after flee: %#v", s.Enemy)
	}
}

func TestTutorialPracticeEncounterCannotBeFled(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "tutorial_grove"
	s.Character.SetStoryFlag("tutorial_available")
	s.Enemy = &combat.Enemy{Name: "Practice Satyr", Level: 1, HP: 20, MaxHP: 28, Damage: 3}

	s.flee()
	if s.Enemy == nil || s.Enemy.HP != 20 {
		t.Fatal("fleeing should not remove or damage the controlled tutorial opponent")
	}
	if s.EncounterClearedRoom == "tutorial_grove" {
		t.Fatal("tutorial encounter should not be marked cleared by fleeing")
	}
}

func TestCombatModeCanToggleOnTheFly(t *testing.T) {
	s := newChoiceTestSession(t)
	if !s.Character.AutoCombat {
		t.Fatal("new characters should start with automatic combat enabled")
	}
	s.handleCommand("pc")
	if s.Character.AutoCombat {
		t.Fatal("pc should toggle automatic combat off")
	}
	s.handleCommand("pc on")
	if !s.Character.AutoCombat {
		t.Fatal("pc on should enable automatic combat")
	}
	s.handleCommand("combatmode off")
	if s.Character.AutoCombat {
		t.Fatal("combatmode off should disable automatic combat")
	}
	s.handleCommand("combatmode toggle")
	if !s.Character.AutoCombat {
		t.Fatal("combatmode toggle should enable automatic combat")
	}
}

func TestAutomaticCombatRoundAttacksAndRetaliates(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.AutoCombat = true
	s.Character.HP = 100
	s.Character.Defense = -100
	s.Enemy = &combat.Enemy{Name: "Training Foe", Level: 1, HP: 500, MaxHP: 500, Damage: 20}
	s.EnemyAggressive = true
	before := s.Enemy.HP
	s.automaticCombatRound()
	if s.Enemy == nil || s.Enemy.HP >= before {
		t.Fatal("automatic combat round should deal player weapon damage")
	}
	if s.Character.HP >= 100 {
		t.Fatal("automatic combat round should include the enemy's response")
	}
}

func TestManualCombatModeSuppressesAutomaticRound(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.AutoCombat = false
	s.Enemy = &combat.Enemy{Name: "Training Foe", Level: 1, HP: 500, MaxHP: 500, Damage: 1}
	s.EnemyAggressive = true
	before := s.Enemy.HP
	s.automaticCombatRound()
	if s.Enemy.HP != before {
		t.Fatal("manual combat mode should not auto-attack")
	}
}

func TestQuicklistAssignListAndCast(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Domain = power.Storm
	s.Character.Level = 5
	s.Character.Mana = 100
	s.Enemy = &combat.Enemy{Name: "Practice Harpy", Level: 1, HP: 500, MaxHP: 500, Damage: 1}
	s.EnemyAggressive = true

	s.handleCommand("ql set 1 Stormspark")
	if s.Character.Quicklist[0] != "Stormspark" { t.Fatalf("slot 1 = %q, want Stormspark", s.Character.Quicklist[0]) }
	beforeHP, beforeMana := s.Enemy.HP, s.Character.Mana
	s.handleCommand("ql 1 harpy")
	if s.Enemy.HP >= beforeHP { t.Fatal("quicklist cast should damage the selected enemy") }
	if s.Character.Mana != beforeMana-8 { t.Fatalf("mana after quick cast = %d, want %d",s.Character.Mana,beforeMana-8) }
}

func TestQuicklistRejectsLockedPowerAndInvalidSlot(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.Domain = power.Storm
	s.Character.Level = 5
	s.handleCommand("ql set 1 Thunderfall")
	if s.Character.Quicklist[0] != "" { t.Fatal("locked power should not be assigned") }
	s.handleCommand("ql set 9 Stormspark")
	if s.Character.Quicklist[0] != "" { t.Fatal("invalid slot should not be assigned") }
}


func TestQuestGoldRewardIsGrantedOnlyOnce(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "therons_forge"
	s.Character.SetStoryFlag("forge_discovered")
	s.Character.SetStoryFlag("crafted_any_gear")
	before := s.Character.Gold

	s.updateQuests()
	afterFirstCompletion := s.Character.Gold
	if afterFirstCompletion != before+17 {
		t.Fatalf("gold after forge quest = %d, want %d", afterFirstCompletion, before+17)
	}
	s.updateQuests()
	if s.Character.Gold != afterFirstCompletion {
		t.Fatalf("rechecking completed quests changed gold to %d; want %d", s.Character.Gold, afterFirstCompletion)
	}
	if !s.Character.HasStoryFlag("quest_completed_forge_remembers") {
		t.Fatal("forge quest completion should be recorded")
	}
}


func TestCharonsMemoryQuestRequiresAnOfferedMemory(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "styx_shore"
	s.talk([]string{"Charon", "memory"})
	if !s.Character.HasStoryFlag("charon_memory_heard") {
		t.Fatal("talking to Charon about memory should unlock the side quest")
	}
	q, ok := quest.Get("charons_memory")
	if !ok {
		t.Fatal("Charon's memory quest is missing from the catalogue")
	}
	if s.Character.Quests[q.ID] >= q.Required {
		t.Fatal("quest should not complete before the player offers a memory")
	}

	s.Character.RoomID = "fields_asphodel"
	s.talk([]string{"Unremembered", "memory"})
	if s.Character.HasStoryFlag("asphodel_memory_returned") {
		t.Fatal("talking about memory should not complete the quest by itself")
	}
	s.offerMemory([]string{"memory", "a", "song"})
	if s.Character.HasStoryFlag("asphodel_memory_returned") {
		t.Fatal("a short fragment should not complete the quest")
	}
	s.offerMemory([]string{"memory", "I", "remember", "someone", "singing", "off-key", "to", "comfort", "me", "when", "I", "feared", "the", "dark"})
	if !s.Character.HasStoryFlag("asphodel_memory_returned") {
		t.Fatal("a detailed offered memory should complete the story thread")
	}
	if s.Character.Quests[q.ID] != q.Required {
		t.Fatalf("quest progress = %d, want %d", s.Character.Quests[q.ID], q.Required)
	}
	if len(s.Character.Memories) == 0 || !strings.Contains(s.Character.Memories[len(s.Character.Memories)-1], "singing off-key") {
		t.Fatal("the offered memory should be recorded in the character's soul memories")
	}
}


func TestGetCoinFromAshAddsInventoryAndRemovesRoomObject(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_foothills"
	s.getObject([]string{"coin", "from", "ash"})
	found := false
	for _, carried := range s.Character.Inventory {
		if strings.EqualFold(carried.Name, "scratched bronze coin") { found = true }
	}
	if !found { t.Fatal("get coin from ash should add the coin to inventory") }
	if !s.Character.HasStoryFlag("found_scratched_coin") { t.Fatal("coin pickup should record its story discovery") }
	room := s.World.GetRoom("olympus_foothills")
	if findRoomObject(room, "coin", "", s.Character) != nil {
		t.Fatal("a character should no longer see a coin after collecting it")
	}
	other := character.New("Otherwalker")
	if findRoomObject(room, "coin", "", other) == nil {
		t.Fatal("a different character should still be able to discover the shared-world coin")
	}
	if got := s.Character.Quests["coin_in_the_ash"]; got != 0 {
		t.Fatalf("coin quest progress after pickup = %d, want 0 until examined", got)
	}
	s.dropItem([]string{"coin"})
	stillCarried := false
	for _, carried := range s.Character.Inventory { if strings.EqualFold(carried.Name, "scratched bronze coin") { stillCarried = true } }
	if !stillCarried { t.Fatal("story-critical coin should not be droppable before the mystery is resolved") }
	s.examineTarget([]string{"coin"})
	if got := s.Character.Quests["coin_in_the_ash"]; got != 1 {
		t.Fatalf("coin quest progress after examination = %d, want 1", got)
	}
}

func TestHarpyIsWaryRatherThanHostileOnSight(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_foothills"
	s.spawnEnemy()
	if s.Enemy == nil || s.Enemy.Behavior != "wary" || s.EnemyAggressive {
		t.Fatal("Harpy should be wary and should not attack merely because the player entered")
	}
}

func TestSatyrIsPassiveAndManticoreIsAggressive(t *testing.T) {
	if got := combat.NewSatyr().Behavior; got != "passive" { t.Fatalf("Satyr behavior = %q, want passive", got) }
	if got := combat.NewManticore().Behavior; got != "aggressive" { t.Fatalf("Manticore behavior = %q, want aggressive", got) }
}


func TestTerminalWrappingIgnoresANSIColorCodes(t *testing.T) {
	colored := "\x1b[31mred text\x1b[0m"
	if got := visibleTerminalWidth(colored); got != len("red text") {
		t.Fatalf("visible width = %d, want %d", got, len("red text"))
	}
	line := "\x1b[36mA long room description should wrap at a word boundary without breaking color codes or words.\x1b[0m"
	wrapped := wrapTerminalLine(line, 28)
	if len(wrapped) < 2 { t.Fatal("long prose line should wrap to terminal width") }
	for _, part := range wrapped {
		if visibleTerminalWidth(part) > 28 {
			t.Fatalf("wrapped line has visible width %d, limit 28: %q", visibleTerminalWidth(part), part)
		}
	}
}


func TestEnteringWaryHarpyRoomDoesNotCauseAutomaticDamage(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_gates"
	s.Character.SetStoryFlag("room_discovered_olympus_foothills")
	s.Character.HP = s.Character.MaxHP
	before := s.Character.HP
	s.move("north")
	if s.Character.RoomID != "olympus_foothills" { t.Fatalf("room = %q, want foothills", s.Character.RoomID) }
	if s.Enemy == nil || s.Enemy.Name != "Harpy" { t.Fatal("Harpy should be present in the foothills") }
	if s.EnemyAggressive { t.Fatal("wary Harpy should not be aggressive on sight") }
	if s.Character.HP != before { t.Fatalf("HP changed on entry: %d -> %d", before, s.Character.HP) }
}


func TestMuseumQuestRequiresExaminingTheArtifact(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "modern_museum"
	s.Character.SetStoryFlag("styx_memory_recovered")
	s.updateQuests()
	if got := s.Character.Quests["museum_echo"]; got != 0 {
		t.Fatalf("museum quest progress on entry = %d, want 0 until artifact is examined", got)
	}
	s.examineTarget([]string{"black-glass funerary mask"})
	if got := s.Character.Quests["museum_echo"]; got != 1 {
		t.Fatalf("museum quest progress after examining artifact = %d, want 1", got)
	}
}


func TestDropAndPickupPreservesItemProperties(t *testing.T) {
	s := newChoiceTestSession(t)
	potion := item.Item{Name:"laurel poultice", Description:"A poultice of crushed laurel leaves.", Tier:item.Common, Kind:"consumable", Healing:20}
	s.Character.Inventory = append(s.Character.Inventory, potion)
	s.dropItem([]string{"laurel", "poultice"})
	for _, carried := range s.Character.Inventory {
		if strings.EqualFold(carried.Name, potion.Name) { t.Fatal("dropped item should leave the inventory") }
	}
	r := s.World.GetRoom(s.Character.RoomID)
	dropped := findRoomObject(r, "laurel poultice", "", s.Character)
	if dropped == nil { t.Fatal("dropped item should become a room object") }
	s.getObject([]string{"laurel", "poultice"})
	found := false
	for _, carried := range s.Character.Inventory {
		if strings.EqualFold(carried.Name, potion.Name) {
			found = true
			if carried.Healing != 20 || carried.Kind != "consumable" || carried.Description != potion.Description {
				t.Fatalf("picked-up item lost its properties: %#v", carried)
			}
		}
	}
	if !found { t.Fatal("picked-up item should return to inventory") }
}


func TestOnlineRosterRegistersRefreshesAndUnregistersPlayers(t *testing.T) {
	s1 := newChoiceTestSession(t)
	s2 := newChoiceTestSession(t)
	s1.Character.Name = "Firstwalker"
	s2.Character.Name = "Secondwalker"
	s1.Character.RoomID = "village_square"
	s2.Character.RoomID = "modern_crossroads"
	s1.registerOnline()
	defer s1.unregisterOnline()
	s2.registerOnline()
	defer s2.unregisterOnline()
	players := s1.onlineSnapshotList()
	foundFirst, foundSecond := false, false
	for _, player := range players {
		if player.Name == "Firstwalker" && player.RoomID == "village_square" { foundFirst = true }
		if player.Name == "Secondwalker" && player.RoomID == "modern_crossroads" { foundSecond = true }
	}
	if !foundFirst || !foundSecond { t.Fatal("online roster should contain both registered players and their rooms") }
	s2.Character.RoomID = "modern_museum"
	s2.refreshOnline()
	for _, player := range s1.onlineSnapshotList() {
		if player.Name == "Secondwalker" && player.RoomID != "modern_museum" {
			t.Fatal("roster location should refresh after movement")
		}
	}
	s2.unregisterOnline()
	for _, player := range s1.onlineSnapshotList() {
		if player.Name == "Secondwalker" { t.Fatal("unregistered player should disappear from who list") }
	}
}


func TestSayDistinguishesGreetingsFromNPCAddresses(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "village_lane"
	if s.isNPCAddress([]string{"hi"}) { t.Fatal("a short greeting should be room speech, not an NPC selector") }
	if !s.isNPCAddress([]string{"damon"}) { t.Fatal("a clear NPC name should route to NPC dialogue") }
}


func TestGetCoinAcceptsCompactContainerSyntax(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_foothills"
	s.getObject([]string{"coin", "ash"})
	if !s.Character.HasStoryFlag("found_scratched_coin") {
		t.Fatal("get coin ash should collect the coin from the ash bed")
	}
}

func TestGetAllCollectsVisibleTakeableObjectsOnly(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "olympus_foothills"
	s.getObject([]string{"all"})
	if !s.Character.HasStoryFlag("object_taken_scratched_bronze_coin") {
		t.Fatal("get all should collect the visible coin")
	}
	if !s.Character.HasStoryFlag("object_taken_bronze_oil_lamp") {
		t.Fatal("get all should collect the takeable lamp")
	}
	if s.Character.HasStoryFlag("object_taken_campfire_ash") {
		t.Fatal("get all should not collect scenery containers")
	}
}


func TestCoinOfferingRequiresExaminationAndCompletesOptionalQuest(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "village_square"
	coin := item.Item{Name:"scratched bronze coin", Description:"A scratched bronze coin.", Tier:item.Common, Kind:"quest"}
	s.Character.Inventory = append(s.Character.Inventory, coin)
	s.Character.SetStoryFlag("found_scratched_coin")
	s.offerItem("coin", "bowl")
	if s.Character.HasStoryFlag("offered_coin_at_well") {
		t.Fatal("the coin should not be offered before the player examines its mark")
	}
	s.Character.SetStoryFlag("examined_scratched_coin")
	s.offerItem("coin", "bowl")
	if !s.Character.HasStoryFlag("offered_coin_at_well") { t.Fatal("the offering should be recorded") }
	if !s.Character.HasStoryFlag("quest_completed_return_the_coin") { t.Fatal("the optional return-the-coin quest should complete") }
	for _, carried := range s.Character.Inventory {
		if strings.EqualFold(carried.Name, "scratched bronze coin") { t.Fatal("the offered coin should leave inventory") }
	}
	beforeXP := s.Character.Experience
	s.offerItem("coin", "bowl")
	if s.Character.Experience != beforeXP { t.Fatal("repeating the offering should not award the quest twice") }
}


func TestClosedCofferHidesClueUntilOpenedAndRead(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "manticore_den"
	r := s.World.GetRoom(s.Character.RoomID)
	if findRoomObject(r, "prayer strip", "", s.Character) != nil {
		t.Fatal("the prayer strip should not be visible while the coffer is closed")
	}
	s.openContainer([]string{"coffer"})
	if !s.Character.HasStoryFlag("opened_tomb_coffer") {
		t.Fatal("opening the coffer should record the discovery")
	}
	if findRoomObject(r, "prayer strip", "", s.Character) == nil {
		t.Fatal("the prayer strip should become visible after opening the coffer")
	}
	s.getObject([]string{"prayer", "strip"})
	s.readObject([]string{"prayer", "strip"})
	if !s.Character.HasStoryFlag("read_charred_prayer") {
		t.Fatal("reading the strip should record the clue")
	}
	if !s.Character.HasStoryFlag("quest_completed_prayer_without_name") {
		t.Fatal("reading the clue should complete The Prayer Without a Name")
	}
	if !s.Character.HasStoryFlag("prayer_witnessed") {
		t.Fatal("quest completion should record the soul witness")
	}
}

func TestClosingCofferHidesUncollectedContentsAgain(t *testing.T) {
	s := newChoiceTestSession(t)
	s.Character.RoomID = "manticore_den"
	r := s.World.GetRoom(s.Character.RoomID)
	s.openContainer([]string{"coffer"})
	s.closeContainer([]string{"coffer"})
	if findRoomObject(r, "prayer strip", "", s.Character) != nil {
		t.Fatal("closing the coffer should hide contents that have not been taken")
	}
}
