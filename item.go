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
//           itemByID does not change after started is set to true!
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

	log.Debugf("Found %s def in session: (%T)%+v", id, defValue, defValue)
	itemDef, ok := defValue.(ItemDef)
	if !ok {
		log.Errorf("session(%s) = (%T)%+v != ItemDef", id, defValue, defValue)
		return nil, false
	}

	return itemDef.Item(s), true
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
		log.Debugf("decoding item obj %s: %+v", id, itemDefObj)
		if len(itemDefObj) != 1 {
			return errors.Errorf("item(%s) has %d entries instead of 1 of %s", id, len(itemDefObj), strings.Join(registeredItemDefNames, "|"))
		}
		var itemDefName string
		var itemDefValue interface{}
		for itemDefName, itemDefValue = range itemDefObj {
			break //only one item
		}

		log.Debugf("item(%s) is \"%s\": %+v", id, itemDefName, itemDefValue)
		itemDefTmpl, ok := itemDefByName[itemDefName]
		if !ok {
			return errors.Errorf("file(%s).item(%s) has unknown type {\"%s\":{...}}", fn, id, itemDefName)
		}
		log.Debugf("parsing into %T...", itemDefTmpl)

		itemDefValuePtr := reflect.New(reflect.TypeOf(itemDefTmpl))
		itemDefJSONValue, _ := json.Marshal(itemDefValue)
		if err := json.Unmarshal(itemDefJSONValue, itemDefValuePtr.Interface()); err != nil {
			return errors.Wrapf(err, "cannot unmarshal definition of item(%s) into %T", id, itemDefTmpl)
		}
		if validator, ok := itemDefValuePtr.Interface().(config.Validator); ok {
			if err := validator.Validate(); err != nil {
				return errors.Wrapf(err, "invalid definition for item(%s)", id)
			}
		}
		itemDef, ok := itemDefValuePtr.Elem().Interface().(ItemDef)
		if !ok {
			return errors.Errorf("wierd! %T is not an item def - should be prevented by registerItemDef()!", itemDefValuePtr.Elem().Interface())
		}
		item := itemDef.StaticItem(id)
		if existingItem, ok := itemByID[id]; ok {
			return errors.Errorf("file %s defines item(%s):%T already defined as item(%s):%T", fn, id, item, id, existingItem)
		}
		itemByID[id] = item
		log.Debugf("File(%s): item(%s):%T", fn, id, item)
	}
	return nil
}
