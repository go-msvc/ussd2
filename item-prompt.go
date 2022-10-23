package ussd

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/jansemmelink/utils2/errors"
)

func init() {
	registerItemDef("prompt", PromptDef{})
}

type PromptDef struct {
	Caption   CaptionDef `json:"caption"`
	Name      string     `json:"name"`
	Type      string     `json:"type" doc:"Type of value defaults to string. This determines validation"`
	valueType reflect.Type
}

var typeByName = map[string]reflect.Type{
	"string": reflect.TypeOf(""),
	"int":    reflect.TypeOf(int(0)),
}

func (def *PromptDef) Validate() error {
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
	if def.Type != "" {
		t, ok := typeByName[def.Type]
		if !ok {
			return errors.Errorf("unknown type(%s)", def.Type)
		}
		def.valueType = t
	} else {
		def.valueType = typeByName["string"]
	}

	return nil
}

func (def PromptDef) StaticItem(id string) Item {
	if err := def.Validate(); err != nil {
		log.Errorf("invalid def (%T)%+v: %+v", def, def, err)
		return FinalDef{Caption: CaptionDef{"en": "Service unavailable"}}.StaticItem(id) //still return an item so the call is easy to use
	}
	return &ussdPrompt{id: id, def: def}
}

func (def PromptDef) Item(s Session) Item {
	if s == nil {
		panic("session is nil")
	}

	//store as new item in the session with uuid
	id := "item_prompt_" + uuid.New().String()
	s.Set(id, map[string]interface{}{"prompt": def})

	//return item that can be used locally, but it will be recreated
	//later from session data if control is first passed back to the user
	return &ussdPrompt{id: id, def: def}
}

//Prompt implements ussd.ItemWithInputHandler
type ussdPrompt struct {
	id  string
	def PromptDef
}

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
	staticItemByID[id] = p
	return p
}

func (p ussdPrompt) ID() string {
	return p.id
}

func (p *ussdPrompt) Render(ctx context.Context) string {
	s := ctx.Value(CtxSession{}).(Session)
	return p.def.Caption.Text(s)
}

type Parser interface {
	Parse(s string) UserError
}

func Type(name string, tmpl Parser) {
	t := reflect.TypeOf(tmpl)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	typeByName[name] = t
}

func (p *ussdPrompt) Process(ctx context.Context, input string) ([]Item, error) {
	s := ctx.Value(CtxSession{}).(Session)

	valuePtr := reflect.New(p.def.valueType)
	if scanner, ok := valuePtr.Interface().(Parser); ok {
		if err := scanner.Parse(input); err != nil {
			//prefix the error to the prompt and repeat the prompt
			def := p.def
			for lang, promptText := range p.def.Caption {
				def.Caption = CaptionDef{
					lang: err.Error(lang) + "\n" + promptText, //todo: err.Error() must be able to translate then substitute or list untranslated format string
				}
			}
			return []Item{&ussdPrompt{id: p.id, def: def}}, nil
		}
		v := valuePtr.Elem().Interface()

		log.Debugf("Prompt(%s) stored %s=(%T)%+v", p.id, p.def.Name, v, v)
		s.Set(p.def.Name, v)
	} else {
		log.Debugf("%T does not implement Parser", valuePtr.Interface())
		log.Debugf("Prompt(%s) stored %s=(%T)%+v", p.id, p.def.Name, input, input)
		s.Set(p.def.Name, input)
	}

	log.Debugf("Prompt(%s) stored %s=%s", p.id, p.def.Name, input)
	return nil, nil
}
