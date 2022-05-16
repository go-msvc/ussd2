package ussd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/vservices/utils/errors"
)

//ussdMenu implements ussd.ItemUsrPrompt
type ussdMenu struct {
	id       string
	title    string
	options  []ussdMenuOption
	rendered bool
}

type ussdMenuOption struct {
	caption   string
	nextItems []Item
}

func Menu(id string, title string) *ussdMenu {
	m := &ussdMenu{
		id:       id,
		title:    title,
		options:  []ussdMenuOption{},
		rendered: false,
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
	m.options = append(m.options, ussdMenuOption{
		caption:   caption,
		nextItems: nextItems, //will be executed in series until the last one, expecting text="" and next="" from others
	})
	return m
}

func (m *ussdMenu) Render(ctx context.Context) string {
	//s := ctx.Value(CtxSession{}).(Session)
	if !m.rendered {
		//first time:
		//substitute values into text
		//todo...

		//break into pages
		//todo...

		m.rendered = true
	}

	//see which page to render
	//todo...

	//todo: set in session menu option map -> next item

	menuPage := m.title
	for n, i := range m.options {
		menuPage += fmt.Sprintf("\n%d. %s", n+1, i.caption)
	}

	//prompt user for input showing this page
	return menuPage
}

func (m *ussdMenu) Process(ctx context.Context, input string) ([]Item, error) {
	if i64, err := strconv.ParseInt(input, 10, 64); err == nil && i64 >= 1 && int(i64) <= len(m.options) {
		nextItems := m.options[i64-1].nextItems
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
