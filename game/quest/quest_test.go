package quest

import "testing"

func TestQuestNumbersAreUniqueAndResolvable(t *testing.T) {
	seen := map[int]string{}
	for _, q := range All() {
		if q.Number < 1 {
			t.Fatalf("quest %q has invalid number %d", q.ID, q.Number)
		}
		if prior, ok := seen[q.Number]; ok {
			t.Fatalf("quest number %d is shared by %q and %q", q.Number, prior, q.ID)
		}
		seen[q.Number] = q.ID
		got, ok := GetByNumber(q.Number)
		if !ok || got.ID != q.ID {
			t.Fatalf("GetByNumber(%d) = %#v, %v; want quest %q", q.Number, got, ok, q.ID)
		}
	}
}

func TestQuestTrackerNumbersAreNotDerivedFromSlicePosition(t *testing.T) {
	q, ok := GetByNumber(15)
	if !ok {
		t.Fatal("expected quest number 14 to resolve")
	}
	if q.ID != "letter_among_feathers" {
		t.Fatalf("quest 15 = %q, want letter_among_feathers", q.ID)
	}
}
