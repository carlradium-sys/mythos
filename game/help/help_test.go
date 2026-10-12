package help

import "testing"

func TestEveryListedHelpTopicExists(t *testing.T) {
	for _, name := range Names() {
		if _, ok := Get(name); !ok {
			t.Errorf("Names lists %q but Get cannot find it", name)
		}
	}
}

func TestCoreTopicsAreDetailed(t *testing.T) {
	for _, name := range []string{"start", "movement", "combat", "guard", "progression", "economy", "world", "status", "invoke", "crafting", "consumables", "tutorial"} {
		topic, ok := Get(name)
		if !ok {
			t.Errorf("missing core help topic %q", name)
			continue
		}
		if len(topic.Text) < 250 {
			t.Errorf("help topic %q is too brief: %d characters", name, len(topic.Text))
		}
	}
}
