package ussd

import "context"

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

func Prompt(id string, text string, name string) *ussdPrompt {
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
