package account

import (
	"path/filepath"
	"testing"
)

func TestCharacterStateSurvivesAccountSaveAndLogin(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "accounts"))
	if err := store.Register("wanderer_1", "my-test-passphrase"); err != nil {
		t.Fatal(err)
	}
	a, err := store.Login("wanderer_1", "my-test-passphrase")
	if err != nil { t.Fatal(err) }
	rec, err := a.NewCharacter("Odysseus")
	if err != nil { t.Fatal(err) }
	rec.Level = 4
	rec.SetStoryFlag("oracle_trust")
	a.Characters = append(a.Characters, *rec)
	if err := store.Save(a); err != nil { t.Fatal(err) }
	loaded, err := store.Login("wanderer_1", "my-test-passphrase")
	if err != nil { t.Fatal(err) }
	if len(loaded.Characters) != 1 { t.Fatalf("character count = %d, want 1", len(loaded.Characters)) }
	got := loaded.Characters[0].Character
	if got.Name != "Odysseus" || got.Level != 4 || !got.HasStoryFlag("oracle_trust") {
		t.Fatalf("saved character state was not restored: %#v", got)
	}
}
