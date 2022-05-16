package ussd

import (
	"sync"

	"bitbucket.org/vservices/utils/errors"
)

//ItemDef can be defined in code, loaded from a file or defined in session data to create an item
type ItemDef interface {
	//the Item() method creates the item
	//before the server start, it is called with s==nil to create static items, e.g. from code/file
	//after start, it must be called with s!=nil to create dynamic items in the current session
	Item(s Session) Item
}

//register how we define a menu, prompt, final, set, ...
func registerItemDef(name string, def ItemDef) {
	itemDefMutex.Lock()
	defer itemDefMutex.Unlock()
	if started {
		panic(errors.Errorf("attempt to define itemDef[%s] after started", name))
	}
	if _, ok := itemDefByName[name]; ok {
		panic(errors.Errorf("duplicate itemDefByName[%s]", name))
	}
	itemDefByName[name] = def
}

var (
	itemDefMutex  sync.Mutex
	itemDefByName = map[string]ItemDef{}
)
