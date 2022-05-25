package ussd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jansemmelink/utils2/errors"
)

func init() {
	registerItemDef("menu", MenuDef{})
}

type MenuDef struct {
	Title   CaptionDef `json:"title"`
	Options []MenuOptionDef
}

func (def MenuDef) Validate() error {
	if err := def.Title.Validate(); err != nil {
		return errors.Wrapf(err, "invalid title")
	}
	if len(def.Options) < 1 {
		return errors.Errorf("%d options (requires 1 or more)", len(def.Options))
	}
	for i, o := range def.Options {
		if err := o.Validate(); err != nil {
			return errors.Wrapf(err, "invalid option[%d]", i)
		}
	}
	return nil
}

type MenuOptionDef struct {
	Caption   CaptionDef   `json:"caption"`
	NextItems NextItemsDef `json:"next"`
}

func (def MenuOptionDef) Validate() error {
	if err := def.Caption.Validate(); err != nil {
		return errors.Wrapf(err, "invalid caption")
	}
	if len(def.NextItems.list) < 1 {
		return errors.Errorf("missing/empty next list")
	}
	return nil
}

func (def MenuDef) StaticItem(id string) Item {
	if err := def.Validate(); err != nil {
		log.Errorf("invalid def (%T)%+v: %+v", def, def, err)
		return FinalDef{Caption: CaptionDef{"en": "Service unavailable"}}.StaticItem(id) //still return an item so the call is easy to use
	}
	return &ussdMenu{id: id, def: def}
}

func (def MenuDef) Item(s Session) Item {
	if s == nil {
		if started {
			panic("creating static menu after stated")
		}
	} else {

	}

	//store as new item in the session with uuid
	id := "_item_menu_" + uuid.New().String()
	s.Set(id, def)

	//return item that can be used locally, but it will be recreated
	//later from session data if control is first passed back to the user
	return &ussdMenu{id: id, def: def}
}

//ussdMenu implements ussd.ItemUsrPrompt
type ussdMenu struct {
	id  string
	def MenuDef
}

func DynMenuDef(title CaptionDef) MenuDef {
	def := MenuDef{Title: title, Options: []MenuOptionDef{}}
	return def
}

func (def MenuDef) With(caption CaptionDef, nextItems ...Item) MenuDef {
	optionDef := MenuOptionDef{
		Caption:   caption,
		NextItems: NextItemsDef{list: []NextItem{}},
	}
	for _, n := range nextItems {
		optionDef.NextItems.list = append(optionDef.NextItems.list, NextItem{ID: n.ID(), Item: n})
	}
	def.Options = append(def.Options, optionDef)
	return def
}

func Menu(id string, title CaptionDef) *ussdMenu {
	if started {
		panic(errors.Errorf("attempt to define static item Menu(%s) after started", id))
	}
	if id == "" {
		panic(errors.Errorf("Menu(%s,%s)", id, title))
	}
	if err := title.Validate(); err != nil {
		panic(errors.Errorf("Menu(%s) with invalid title: %+v", id, err))
	}
	m := &ussdMenu{
		id: id,
		def: MenuDef{
			Title:   title,
			Options: []MenuOptionDef{},
		},
	}
	staticItemByID[id] = m
	return m
}

func (m ussdMenu) ID() string { return m.id }

func (m *ussdMenu) With(caption CaptionDef, nextItems ...Item) *ussdMenu {
	if len(nextItems) > 0 { //if menu item is implemented, nextItems may not be nil
		for i := 0; i < len(nextItems); i++ {
			if nextItems[i] == nil {
				panic(fmt.Sprintf("menu(%s).With(%s).next[%d]==nil", m.id, caption, i))
			}
		}
	}
	optionDef := MenuOptionDef{
		Caption:   caption,
		NextItems: NextItemsDef{list: []NextItem{}}, //will be executed in series until the last one, expecting text="" and next="" from others
	}
	for _, n := range nextItems {
		optionDef.NextItems.list = append(optionDef.NextItems.list, NextItem{ID: n.ID(), Item: n})
	}
	m.def.Options = append(m.def.Options, optionDef)
	return m
}

func (m *ussdMenu) Render(ctx context.Context) string {
	s := ctx.Value(CtxSession{}).(Session)
	//time.Sleep(time.Second) //todo: remove, was just to test console

	//see which page to render
	//todo...

	//todo: set in session menu option map -> next item
	menuPage := m.def.Title.Text(s)
	for i, o := range m.def.Options {
		menuPage += fmt.Sprintf("\n%d. %s", i+1, o.Caption.Text(s))
	}

	//prompt user for input showing this page
	return menuPage
}

func (m *ussdMenu) Process(ctx context.Context, input string) ([]Item, error) {
	s := ctx.Value(CtxSession{}).(Session)
	if i64, err := strconv.ParseInt(input, 10, 64); err == nil && i64 >= 1 && int(i64) <= len(m.def.Options) {
		nextItems := m.def.Options[i64-1].NextItems
		if len(nextItems.list) == 0 {
			log.Errorf("menu(%s) input(%s): this item is not yet implemented", m.id, input)
			return nil, errors.Errorf("not yet implemented")
		}

		log.Debugf("menu(%s) selected(%s) -> next: %s", m.id, input, strings.Join(nextItems.Ids(), ","))
		//todo: generic resolve of next items - some/all may already be resolved when we get here
		items, err := nextItems.Items(s)
		if err != nil {
			return nil, errors.Wrapf(err, "some item ids cannot be resolved")
		}
		return items, nil
	}
	log.Debugf("menu(%s) input(%s) unknown - display same menu again", m.id, input)
	return []Item{m}, nil //redisplay this same menu without error
}
