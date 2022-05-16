package ussd

import "context"

func Final(id string, text string) *ussdFinal {
	f := &ussdFinal{
		id:   id,
		text: text,
	}
	itemByID[id] = f
	return f
}

//Final implements ussd.Item
type ussdFinal struct {
	id   string
	text string
}

func (f ussdFinal) ID() string { return f.id }

func (f ussdFinal) Render(ctx context.Context) string {
	return f.text
}
