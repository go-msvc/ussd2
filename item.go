package ussd

import "context"

//Item is any type of USSD service processing step
type Item interface {
	ID() string
}

type ItemSvc interface {
	Item
	Exec(ctx context.Context) (nextItems []Item, err error) //err to stop
}

// type ItemSvcWait interface {
// 	Item
// 	Request(ctx context.Context) (err error)                    //err to stop
// 	Process(ctx context.Context, value interface{}) (err error) //err to stop
// }

type ItemUsr interface {
	Item
	Render(ctx context.Context) string
}

type ItemUsrPrompt interface {
	ItemUsr
	Process(ctx context.Context, input string) (nextItems []Item, err error) //return self to repeat prompt, err to display to user
}

//all static items (defined at startup for this release) must have an id
//so that sessions can continue on any instance of the service with the same release
//
//items may be registered by ID during startup only so that in a production
//environment where the service and its config is defined in a container, the
//same list of items will exist in each instance of the service.
//
//After startup, dynamic items can still be created, but they will be defined in the session
//and be deleted when the session ends, so they will be limited to things like set(name=value)
//where the value is session specific and stored in the session for each set item
//
//IMPORTANT: dynamic items are not stored in itemByID[]!!!
//           itemByID does not change after started is set to true!
var (
	itemByID = map[string]Item{}
	started  = false
)

//ItemDef is an item definition that can be loaded from file or session data to create an item
type ItemDef interface {
	Item(s Session) Item
}

func ItemByID(id string, s Session) (Item, bool) {
	//see if static item:
	if item, ok := itemByID[id]; ok {
		return item, true //found static item
	}
	if s == nil {
		return nil, false //no static item and not currently in a session
	}
	defValue := s.Get(id)
	if defValue == nil {
		return nil, false //also not present in the current session
	}

	log.Debugf("Found %s def in session: (%T)%+v", id, defValue, defValue)
	itemDef, ok := defValue.(ItemDef)
	if !ok {
		log.Errorf("session(%s) = (%T)%+v != ItemDef", id, defValue, defValue)
		return nil, false
	}

	return itemDef.Item(s), true
}
