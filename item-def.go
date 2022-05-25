package ussd

import (
	"sync"

	"github.com/jansemmelink/utils2/errors"
)

//ItemDef can be defined in code, loaded from a file or defined in session data to create an item
type ItemDef interface {
	//the StaticItem() creates a static item before the first transaction is started
	//it panics if called after startup
	StaticItem(id string) Item
	//the Item() method creates a dynamic item at runtime
	//it panics if called during startup (without a session)
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
	registeredItemDefNames = append(registeredItemDefNames, name)
}

var (
	itemDefMutex           sync.Mutex
	itemDefByName          = map[string]ItemDef{}
	registeredItemDefNames = []string{}
)
