package ussd

import (
	"context"

	"bitbucket.org/vservices/utils/errors"
	"github.com/google/uuid"
)

type FinalDef struct {
	Text string `json:"text"`
}

func (def FinalDef) Validate() error {
	if def.Text == "" {
		return errors.Errorf("missing text")
	}
	return nil
}

func DynFinal(s Session, text string) *ussdFinal {
	if s == nil {
		panic("session is nil")
	}
	def := FinalDef{Text: text}
	if err := def.Validate(); err != nil {
		log.Errorf("invalid final def %+v: %+v", def, err)
		return DynFinal(s, "Service unavailable") //still return an item so the call is easy to use
	}
	//store as new item in the session with uuid
	id := "_item_final_" + uuid.New().String()
	s.Set(id, def)

	//return final that can be used locally, but it will be recreated
	//later from session data if control is first passed back to the user
	return &ussdFinal{id: id, def: def}
}

func Final(id string, text string) *ussdFinal {
	if started {
		panic(errors.Errorf("attempt to define static item Final(%s) after started", id))
	}
	def := FinalDef{Text: text}
	if id == "" {
		panic(errors.Errorf("Final(%s)", id))
	}
	if err := def.Validate(); err != nil {
		panic(errors.Wrapf(err, "invalid final def %+v: %+v", def))
	}
	if existingItem, ok := itemByID[id]; ok {
		if existingFinalItem, ok := existingItem.(*ussdFinal); ok {
			if existingFinalItem.def.Text != def.Text {
				panic(errors.Errorf("Final(%s) redefined with different text(%s != %s)", id, def.Text, existingFinalItem.def.Text))
			}
			return existingFinalItem
		}
		panic(errors.Errorf("Final(%s) redefined %T(%s)", id, existingItem, id))
	}
	f := &ussdFinal{
		id:  id,
		def: def,
	}
	itemByID[id] = f
	return f
}

//Final implements ussd.Item
type ussdFinal struct {
	id  string
	def FinalDef
}

func (f ussdFinal) ID() string { return f.id }

func (f ussdFinal) Render(ctx context.Context) string {
	return f.def.Text
}
