package ussd

import (
	"context"

	"bitbucket.org/vservices/utils/errors"
)

func init() {
	registerItemDef("prompt", PromptDef{})
}

type PromptDef struct {
	Caption CaptionDef `json:"caption"`
	Name    string     `json:"name"`
}

func (def PromptDef) Validate() error {
	if err := def.Caption.Validate(); err != nil {
		return errors.Wrapf(err, "invalid caption")
	}
	if def.Name == "" {
		return errors.Errorf("missing name")
	}
	if !snakeCaseRegex.MatchString(def.Name) {
		return errors.Errorf("name:\"%s\" is not written in snake_case", def.Name)
	}
	if def.Name[0] == '_' {
		return errors.Errorf("name:\"%s\" may not start with '_'", def.Name) //reserved for dynamic items
	}
	return nil
}

func (def PromptDef) Item(s Session) Item {
	panic("NYI")
}

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
