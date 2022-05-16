package ussd

import (
	"context"

	"bitbucket.org/vservices/utils/errors"
	"github.com/google/uuid"
)

type SetDef struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value,omitempty"` //todo: should support expressions
}

func (def SetDef) Validate() error {
	if def.Name == "" {
		return errors.Errorf("missing name")
	}
	return nil
}

func DynSet(s Session, name string, value interface{}) Item {
	if s == nil {
		panic("session is nil")
	}
	def := SetDef{Name: name, Value: value}
	if err := def.Validate(); err != nil {
		log.Errorf("invalid set def %+v: %+v", def, err)
		return DynFinal(s, "Service unavailable") //still return an item so the call is easy to use
	}

	//store as new item in the session with uuid
	id := "_item_set_" + uuid.New().String()
	s.Set(id, def)

	//return set that can be used locally, but it will be recreated
	//later from session data if control is first passed back to the user
	return ussdSet{id: id, def: def}
}

//static set item
func Set(name string, value interface{}) Item {
	if started {
		panic(errors.Errorf("attempt to define static item Set(%s=%v) after started", name, value))
	}
	def := SetDef{Name: name, Value: value}
	if err := def.Validate(); err != nil {
		panic(errors.Wrapf(err, "invalid set def %+v: %+v", def))
	}

	id := uuid.New().String()
	// if existingItem, ok := ItemByID(id); ok {
	// 	existingSetItem, ok := existingItem.(ussdSet)
	// 	if !ok {
	// 		panic(errors.Errorf("Set(%s) redefines existing item(%s) with type %T", id, id, existingSetItem))
	// 	}
	// 	if existingSetItem.def.Name != def.Name {
	// 		panic(errors.Errorf("Set(%s).name=\"%s\" redefines a different name Set(%s).name=\"%s\"", id, def.Name, id, existingSetItem.def.Name))
	// 	}
	// 	if existingSetItem.def.Value != def.Value {
	// 		panic(errors.Errorf("Set(%s).%s=(%T)%v redefines a different value Set(%s).%s=(%T)%v", id, def.Name, def.Value, def.Value, id, existingSetItem.def.Name, existingSetItem.def.Value, existingSetItem.def.Value))
	// 	}
	// 	return existingSetItem //ok to reuse
	// }

	//create a new set item
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
