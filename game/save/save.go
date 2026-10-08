package save

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"

	"fatewalker/game/character"
)

var validSlot = regexp.MustCompile("^[A-Za-z0-9_-]{1,32}$")

func path(slot string) (string, error) {
	if !validSlot.MatchString(slot) {
		return "", errors.New("invalid save slot")
	}
	return filepath.Join("saves", slot+".json"), nil
}

func Save(slot string, c *character.Character) error {
	p, err := path(slot)
	if err != nil { return err }
	if err := os.MkdirAll("saves", 0755); err != nil { return err }
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil { return err }
	return os.WriteFile(p, data, 0600)
}

func Load(slot string) (*character.Character, error) {
	p, err := path(slot)
	if err != nil { return nil, err }
	data, err := os.ReadFile(p)
	if err != nil { return nil, err }
	var c character.Character
	if err := json.Unmarshal(data, &c); err != nil { return nil, err }
	if c.Name == "" || c.Level < 1 || c.Life < 1 {
		return nil, errors.New("save data is invalid")
	}
	return &c, nil
}
