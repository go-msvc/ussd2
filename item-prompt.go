package ussd

import (
	"context"

	"github.com/google/uuid"
	"github.com/jansemmelink/utils2/errors"
)

func init() {
	registerItemDef("prompt", PromptDef{})
}

type PromptDef struct {
	Caption CaptionDef `json:"caption"`
	Name    string     `json:"name"`
	//validators []InputValidator???
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
	if s == nil {
		panic("session is nil")
	}

	//store as new item in the session with uuid
	id := "_item_prompt_" + uuid.New().String()
	s.Set(id, def)

	//return item that can be used locally, but it will be recreated
	//later from session data if control is first passed back to the user
	return &ussdPrompt{id: id, def: def}
}

//Prompt implements ussd.ItemWithInputHandler
type ussdPrompt struct {
	id  string
	def PromptDef
}

type InputValidator interface {
	Validate(input string) error
}

// func DynPrompt(id string, def PromptDef) Item {
// 	//todo...
// }

func Prompt(id string, caption CaptionDef, name string) *ussdPrompt {
	if started {
		panic(errors.Errorf("attempt to define static item Prompt(%s) after started", id))
	}
	if id == "" || !snakeCaseRegex.MatchString(name) {
		panic(errors.Errorf("Prompt(id:%s,name:%s)", id, name))
	}
	if err := caption.Validate(); err != nil {
		panic(errors.Wrapf(err, "invalid prompt caption"))
	}
	p := &ussdPrompt{
		id:  id,
		def: PromptDef{Caption: caption, Name: name},
	}
	itemByID[id] = p
	return p
}

func (p ussdPrompt) ID() string {
	return p.id
}

func (p *ussdPrompt) Render(ctx context.Context) string {
	s := ctx.Value(CtxSession{}).(Session)
	return p.def.Caption.Text(s)
}

func (p *ussdPrompt) Process(ctx context.Context, input string) ([]Item, error) {
	s := ctx.Value(CtxSession{}).(Session)
	// for _, v := range p.validators {
	// 	if err := v.Validate(input); err != nil {
	// 		return []Item{p}, err //repeat prompt with error message
	// 	}
	// }
	//todo: optional validator + invalid message
	s.Set(p.def.Name, input)
	log.Debugf("Prompt(%s) stored %s=%s", p.id, p.def.Name, input)
	return nil, nil
}
