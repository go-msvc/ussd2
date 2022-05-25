package ussd

import (
	"context"

	"github.com/google/uuid"
	"github.com/jansemmelink/utils2/errors"
)

func init() {
	registerItemDef("final", FinalDef{})
}

type FinalDef struct {
	Caption CaptionDef `json:"caption"`
}

func (def FinalDef) Validate() error {
	if err := def.Caption.Validate(); err != nil {
		return errors.Wrapf(err, "invalid caption")
	}
	return nil
}

func (def FinalDef) StaticItem(id string) Item {
	if err := def.Validate(); err != nil {
		log.Errorf("invalid def (%T)%+v: %+v", def, def, err)
		return FinalDef{Caption: CaptionDef{"en": "Service unavailable"}}.StaticItem(id) //still return an item so the call is easy to use
	}
	return ussdFinal{id: id, def: def}
}

func (def FinalDef) Item(s Session) Item {
	if s == nil {
		panic("session is nil")
	}
	if err := def.Validate(); err != nil {
		log.Errorf("invalid def (%T)%+v: %+v", def, def, err)
		return FinalDef{Caption: CaptionDef{"en": "Service unavailable"}}.Item(s) //still return an item so the call is easy to use
	}
	//store as new item in the session with uuid
	id := "_item_final_" + uuid.New().String()
	s.Set(id, def)

	//return final that can be used locally, but it will be recreated
	//later from session data if control is first passed back to the user
	return ussdFinal{id: id, def: def}
}

//define a static final response
func Final(id string, caption CaptionDef) *ussdFinal {
	if started {
		panic(errors.Errorf("attempt to define static item Final(%s) after started", id))
	}
	def := FinalDef{Caption: caption}
	if err := def.Validate(); err != nil {
		panic(errors.Wrapf(err, "invalid final def %+v: %+v", def))
	}
	if id == "" {
		id = uuid.New().String()
	} else {
		if existingItem, ok := staticItemByID[id]; ok {
			panic(errors.Errorf("Final(%s) redefines %T(%s)", id, existingItem, id))
		}
	}
	f := &ussdFinal{
		id:  id,
		def: def,
	}
	staticItemByID[id] = f
	return f
}

//Final implements ussd.Item
type ussdFinal struct {
	id  string
	def FinalDef
}

func (f ussdFinal) ID() string { return f.id }

func (f ussdFinal) Render(ctx context.Context) string {
	return f.def.Caption["fr"] //todo: use lang code for session
}
