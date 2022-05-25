package ussd

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"

	"github.com/jansemmelink/utils2/config"
	"github.com/jansemmelink/utils2/errors"
)

//Item is any type of USSD service processing step
type Item interface {
	ID() string
}

type ItemSvc interface {
	Item
	Exec(ctx context.Context) (nextItems []Item, err error) //err to stop
}

type ItemUsr interface {
	Item
	Render(ctx context.Context) string
}

type ItemUsrPrompt interface {
	ItemUsr
	Process(ctx context.Context, input string) (nextItems []Item, err error) //return self to repeat prompt, err will be displayed to user and will end the session
}

//all static items (defined at startup for this release) must have an id
//so that sessions can continue on any instance of the service with the same release
//
//items may only be registered by ID during startup, so that in a production
//environment where the service and its config is defined in a container, the
//same list of items will exist in each instance of the service of the same version.
//
//After startup, dynamic items can still be created, but they will be defined in the session
//and be deleted when the session ends, so they will be limited to things like set(name=value)
//where the value is session specific and stored in the session for each set item
//
//IMPORTANT: dynamic items are not stored in staticItemByID[]!!!
//           staticItemByID does not change after started is set to true!
var (
	staticItemByID = map[string]Item{}
	started        = false
)

func ItemByID(id string, s Session) (Item, bool) {
	//see if static item:
	if item, ok := staticItemByID[id]; ok {
		return item, true //found static item
	}
	if s == nil {
		return nil, false //no static item and not currently in a session
	}
	defValue := s.Get(id)
	if defValue == nil {
		return nil, false //also not present in the current session
	}

	itemDefObj, ok := defValue.(map[string]interface{}) //json object read from session data
	if !ok {
		log.Errorf("session(%s) = (%T)%+v != ItemDefObj", id, defValue, defValue)
		return nil, false
	}
	item, err := makeItem(id, itemDefObj)
	if err != nil {
		log.Errorf("failed to make item(%s) from session object %+v: %+v", id, itemDefObj, err)
		return nil, false
	}
	return item, true
}

func LoadItems(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return errors.Wrapf(err, "cannot open file %s", fn)
	}
	defer f.Close()
	var itemDefsInFile map[string]map[string]interface{}
	if err := json.NewDecoder(f).Decode(&itemDefsInFile); err != nil {
		return errors.Wrapf(err, "failed to decode items from file %s", fn)
	}

	for id, itemDefObj := range itemDefsInFile {
		if len(itemDefObj) != 1 {
			return errors.Errorf("item(%s) has %d entries instead of 1 of %s", id, len(itemDefObj), strings.Join(registeredItemDefNames, "|"))
		}
		item, err := makeItem(id, itemDefObj)
		if err != nil {
			return errors.Wrapf(err, "failed to make item")
		}
		if existingItem, ok := staticItemByID[id]; ok {
			return errors.Errorf("file %s defines item(%s):%T already defined as item(%s):%T", fn, id, item, id, existingItem)
		}
		staticItemByID[id] = item
	}
	return nil
}

func makeItem(id string, itemDefObj map[string]interface{}) (Item, error) {
	var itemDefName string
	var itemDefValue interface{}
	for itemDefName, itemDefValue = range itemDefObj {
		break //only one item
	}

	itemDefTmpl, ok := itemDefByName[itemDefName]
	if !ok {
		return nil, errors.Errorf("item(%s) has unknown type {\"%s\":{...}}", id, itemDefName)
	}

	itemDefValuePtr := reflect.New(reflect.TypeOf(itemDefTmpl))
	itemDefJSONValue, _ := json.Marshal(itemDefValue)
	if err := json.Unmarshal(itemDefJSONValue, itemDefValuePtr.Interface()); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal definition of item(%s) into %T", id, itemDefTmpl)
	}
	if validator, ok := itemDefValuePtr.Interface().(config.Validator); ok {
		if err := validator.Validate(); err != nil {
			return nil, errors.Wrapf(err, "invalid definition for item(%s)", id)
		}
	}
	itemDef, ok := itemDefValuePtr.Elem().Interface().(ItemDef)
	if !ok {
		return nil, errors.Errorf("wierd! %T is not an item def - should be prevented by registerItemDef()!", itemDefValuePtr.Elem().Interface())
	}
	return itemDef.StaticItem(id), nil
}
