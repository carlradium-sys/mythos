package session

import (
	"io"
	"net"
	"strings"
	"testing"

	"fatewalker/game/character"
	"fatewalker/game/combat"
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
	if s.Character.Experience != 120 {
		t.Fatalf("experience = %d, want 120", s.Character.Experience)
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
	n := s.currentNPC()
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
