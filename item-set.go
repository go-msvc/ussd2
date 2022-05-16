package ussd

import (
	"context"
	"regexp"

	"bitbucket.org/vservices/utils/errors"
	"github.com/google/uuid"
)

func init() {
	registerItemDef("set", SetDef{})
}

type SetDef struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value,omitempty"` //todo: should support expressions
}

func (def SetDef) Validate() error {
	if def.Name == "" {
		return errors.Errorf("missing name")
	}
	if !snakeCaseRegex.MatchString(def.Name) {
		return errors.Errorf("name:\"%s\" is not written in snake_case", def.Name)
	}
	if def.Name[0] == '_' {
		return errors.Errorf("name:\"%s\" may not start with '_'", def.Name) //reserved for dynamic items
	}
	//value may be practically anything, even nil to delete session data, so not validated
	return nil
}

//creates a dynamic item in the session
func (def SetDef) Item(s Session) Item {
	if s == nil {
		panic("session is nil")
	}
	if err := def.Validate(); err != nil {
		log.Errorf("invalid def (%T)%+v: %+v", def, def, err)
		return FinalDef{Caption: CaptionDef{"en": "Service unavailable"}}.Item(s) //still return an item so the call is easy to use
	}
	//store as new item in the session with uuid
	id := "_item_set_" + uuid.New().String()
	s.Set(id, def)

	//return set that can be used locally, but it will be recreated
	//later from session data if control is first passed back to the user
	return ussdSet{id: id, def: def}
} //SetDef.Item()

//define a static set(name=value) item
func Set(id string, name string, value interface{}) Item {
	if started {
		panic(errors.Errorf("attempt to define static item Set(%s=%v) after started", name, value))
	}
	def := SetDef{Name: name, Value: value}
	if err := def.Validate(); err != nil {
		panic(errors.Wrapf(err, "invalid set def %+v: %+v", def))
	}
	if id == "" {
		id = uuid.New().String()
	} else {
		if existingItem, ok := itemByID[id]; ok {
			panic(errors.Errorf("Set(%s) redefines %T(%s)", id, existingItem, id))
		}
	}
	s := ussdSet{
		id:  id,
		def: def,
	}
	itemByID[id] = s
	return s
}

type ussdSet struct {
	id  string
	def SetDef
}

func (set ussdSet) ID() string {
	return set.id
}

func (set ussdSet) Exec(ctx context.Context) ([]Item, error) {
	s := ctx.Value(CtxSession{}).(Session)
	s.Set(set.def.Name, set.def.Value)
	return nil, nil
}

const snakeCasePattern = `[a-z]([a-z0-9_]*[a-z0-9])*`

var snakeCaseRegex = regexp.MustCompile("^" + snakeCasePattern + "$")
