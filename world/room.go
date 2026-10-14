package world

import "sync"

type Room struct {
	ID          string
	Name        string
	Description string
	Exits       map[string]string
	Objects     []*Object
	objectsMu sync.RWMutex
}

func NewRoom(id, name, description string) *Room {
	return &Room{
		ID:          id,
		Name:        name,
		Description: description,
		Exits:       make(map[string]string),
		Objects:     []*Object{},
	}
}


// ObjectsSnapshot returns a stable slice for concurrent session readers. Room
// objects are shared by all players, so runtime additions must use AddObject.
func (r *Room) ObjectsSnapshot() []*Object {
	if r == nil {
		return nil
	}
	r.objectsMu.RLock()
	defer r.objectsMu.RUnlock()
	return append([]*Object(nil), r.Objects...)
}

// AddObject safely adds a runtime object to a shared room.
func (r *Room) AddObject(object *Object) {
	if r == nil || object == nil {
		return
	}
	r.objectsMu.Lock()
	r.Objects = append(r.Objects, object)
	r.objectsMu.Unlock()
}


// RemoveObjectByID atomically removes a runtime object, allowing shared ground
// items to be picked up by at most one player.
func (r *Room) RemoveObjectByID(id string) (*Object, bool) {
	if r == nil || id == "" {
		return nil, false
	}
	r.objectsMu.Lock()
	defer r.objectsMu.Unlock()
	for i, object := range r.Objects {
		if object != nil && object.ID == id {
			r.Objects = append(r.Objects[:i], r.Objects[i+1:]...)
			return object, true
		}
	}
	return nil, false
}
