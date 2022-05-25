package ussd

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jansemmelink/utils2/errors"
)

type NextItemsDef struct {
	list []NextItem
}

//returns error if some IDs cannot be resolved
func (nid NextItemsDef) Items(s Session) ([]Item, error) {
	items := make([]Item, len(nid.list))
	for index, nextItem := range nid.list {
		if nextItem.Item == nil {
			var ok bool
			nextItem.Item, ok = ItemByID(nextItem.ID, s)
			if !ok {
				return nil, errors.Errorf("next[%d]=\"%s\" not found", index, nextItem.ID)
			}
			nid.list[index] = nextItem //now resolved
		}
		items[index] = nextItem.Item
	}
	return items, nil //all are resolved
}

func (nid NextItemsDef) Ids() []string {
	ids := make([]string, len(nid.list))
	for i, n := range nid.list {
		ids[i] = n.ID
	}
	return ids
}

func (nid *NextItemsDef) UnmarshalJSON(value []byte) error {
	//expect value to be a JSON list of strings or objects or a mix of both:
	//    ["a","b"]
	//or
	//    ["a", {"menu":{...}}]
	//the objects will be created as non-reuasable items using id=uuid()
	log.Debugf("JSON Unmarshal: %s", string(value))

	list := []interface{}{}
	if err := json.Unmarshal(value, &list); err != nil {
		return errors.Wrapf(err, "cannot unmarshal list of items")
	}
	log.Debugf("got %d next items", len(list))

	nid.list = []NextItem{}
	for index, nextDef := range list {
		log.Debugf("  next[%d]: (%T)%v", index, nextDef, nextDef)
		switch nextDef := nextDef.(type) {
		case string:
			//expect this to be an item ID that will later be resolved
			if nextDef == "" {
				return errors.Errorf("next[%d] is empty", index)
			}
			if !snakeCaseRegex.MatchString(nextDef) {
				return errors.Errorf("next[%d]:\"%s\" is not snake_case", index, nextDef)
			}
			nid.list = append(nid.list, NextItem{ID: nextDef, Item: nil})
		case map[string]interface{}:
			if len(nextDef) != 1 {
				return errors.Errorf("next[%d]:%+v must have one element, not %d", index, nextDef, len(nextDef))
			}
			id := uuid.New().String() //only used here, so id not significant
			item, err := makeItem(id, nextDef)
			if err != nil {
				return errors.Wrapf(err, "failed to make item(%s)", id)
			}
			staticItemByID[id] = item
			nid.list = append(nid.list, NextItem{ID: id, Item: item})
		default:
			return errors.Errorf("cannot unmarshal next item[%d] from %T", index, nextDef)
		}
	}
	return nil
}

func (nid NextItemsDef) MarshalJSON() ([]byte, error) {
	return nil, errors.Errorf("NYI")
}

//NextItem stores at least the ID and when defined/resolved, also the item
type NextItem struct {
	ID   string
	Item Item
}
