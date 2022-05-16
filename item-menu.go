package ussd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/vservices/utils/errors"
	"github.com/google/uuid"
)

//ussdMenu implements ussd.ItemUsrPrompt
type ussdMenu struct {
	id  string
	def MenuDef
}

type MenuDef struct {
	Title   string `json:"title"`
	Options []MenuOptionDef
}

type MenuOptionDef struct {
	Caption   string `json:"caption"`
	NextItems []Item `json:"next_items"`
}

func DynMenuDef(title string) MenuDef {
	def := MenuDef{Title: title, Options: []MenuOptionDef{}}
	return def
}

func (def MenuDef) With(caption string, nextItems ...Item) MenuDef {
	def.Options = append(def.Options, MenuOptionDef{
		Caption:   caption,
		NextItems: nextItems,
	})
	return def
}

func (def MenuDef) Item(s Session) Item {
	if s == nil {
		panic("session is nil")
	}

	//store as new item in the session with uuid
	id := "_item_menu_" + uuid.New().String()
	s.Set(id, def)

	//return item that can be used locally, but it will be recreated
	//later from session data if control is first passed back to the user
	return &ussdMenu{id: id, def: def}
}

func Menu(id string, title string) *ussdMenu {
	if started {
		panic(errors.Errorf("attempt to define static item Menu(%s) after started", id))
	}
	if id == "" || title == "" {
		panic(errors.Errorf("Menu(%s,%s)", id, title))
	}
	m := &ussdMenu{
		id: id,
		def: MenuDef{
			Title:   title,
			Options: []MenuOptionDef{},
		},
	}
	itemByID[id] = m
	return m
}

func (m ussdMenu) ID() string { return m.id }

func (m *ussdMenu) With(caption string, nextItems ...Item) *ussdMenu {
	if len(nextItems) > 0 { //if menu item is implemented, nextItems may not be nil
		for i := 0; i < len(nextItems); i++ {
			if nextItems[i] == nil {
				panic(fmt.Sprintf("menu(%s).With(%s).next[%d]==nil", m.id, caption, i))
			}
		}
	}
	m.def.Options = append(m.def.Options, MenuOptionDef{
		Caption:   caption,
		NextItems: nextItems, //will be executed in series until the last one, expecting text="" and next="" from others
	})
	return m
}

func (m *ussdMenu) Render(ctx context.Context) string {
	//s := ctx.Value(CtxSession{}).(Session)
	//time.Sleep(time.Second) //todo: remove, was just to test console

	//see which page to render
	//todo...

	//todo: set in session menu option map -> next item

	menuPage := m.def.Title
	for i, o := range m.def.Options {
		menuPage += fmt.Sprintf("\n%d. %s", i+1, o.Caption)
	}

	//prompt user for input showing this page
	return menuPage
}

func (m *ussdMenu) Process(ctx context.Context, input string) ([]Item, error) {
	if i64, err := strconv.ParseInt(input, 10, 64); err == nil && i64 >= 1 && int(i64) <= len(m.def.Options) {
		nextItems := m.def.Options[i64-1].NextItems
		if len(nextItems) == 0 {
			log.Errorf("menu(%s) input(%s): this item is not yet implemented", m.id, input)
			return nil, errors.Errorf("not yet implemented")
		}
		nextIds := []string{}
		for _, i := range nextItems {
			nextIds = append(nextIds, i.ID())
		}
		log.Debugf("menu(%s) selected(%s) -> next: %s", m.id, input, strings.Join(nextIds, ","))
		return nextItems, nil
	}
	log.Debugf("menu(%s) input(%s) unknown - display same menu again", m.id, input)
	return []Item{m}, nil //redisplay this same menu without error
}
