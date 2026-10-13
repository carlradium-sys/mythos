package world

import (
	"strings"
	"fatewalker/game/item"
)

// Object is a physical thing in a room. ContainerID names the visible room
// object that holds it; an empty ContainerID means it rests directly in the room.
type Object struct {
	ID          string
	Name        string
	Description string
	Keywords    []string
	ContainerID string
	Visible     bool
	Takeable    bool
	Container   bool
	StoredItem  *item.Item
}

func NewObject(id, name, description string, keywords ...string) *Object {
	return &Object{ID:id, Name:name, Description:description, Keywords:keywords, Visible:true}
}

func (o *Object) Matches(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if o == nil || q == "" { return false }
	if strings.EqualFold(o.Name, q) || strings.EqualFold(o.ID, q) { return true }
	if strings.Contains(strings.ToLower(o.Name), q) { return true }
	for _, k := range o.Keywords {
		if strings.EqualFold(k, q) || strings.Contains(strings.ToLower(k), q) { return true }
	}
	return false
}

// VisibleIn reports whether an object can be discovered without first opening
// or searching its parent container. The ash-covered coin is deliberately
// discoverable: careful observation reveals its outline in the grey ash.
func (o *Object) VisibleIn(objects []*Object) bool {
	if o == nil || !o.Visible { return false }
	if o.ContainerID == "" { return true }
	for _, parent := range objects {
		if parent != nil && parent.ID == o.ContainerID {
			return parent.Visible && parent.Container
		}
	}
	return false
}
