package ussd

import (
	"context"

	"bitbucket.org/vservices/utils/errors"
)

//Prompt implements ussd.ItemWithInputHandler
type ussdPrompt struct {
	id         string
	text       string
	name       string
	validators []InputValidator
}

type InputValidator interface {
	Validate(input string) error
}

// func DynPrompt(id string, def PromptDef) Item {
// 	//todo...
// }

func Prompt(id string, text string, name string) *ussdPrompt {
	if started {
		panic(errors.Errorf("attempt to define static item Prompt(%s) after started", id))
	}
	if id == "" || text == "" || name == "" {
		panic(errors.Errorf("Prompt(%s,%s)", id, text, name))
	}
	p := &ussdPrompt{
		id:         id,
		text:       text,
		name:       name,
		validators: nil,
	}
	itemByID[id] = p
	return p
}

func (p ussdPrompt) ID() string {
	return p.id
}

func (p *ussdPrompt) Render(ctx context.Context) string {
	return p.text
}

func (p *ussdPrompt) Process(ctx context.Context, input string) ([]Item, error) {
	s := ctx.Value(CtxSession{}).(Session)
	for _, v := range p.validators {
		if err := v.Validate(input); err != nil {
			return []Item{p}, err //repeat prompt with error message
		}
	}
	//todo: optional validator + invalid message
	s.Set(p.name, input)
	log.Debugf("Prompt(%s) stored %s=%s", p.id, p.name, input)
	return nil, nil
}
