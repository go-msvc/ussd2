package ussd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/jansemmelink/utils2/errors"
)

type UserError interface {
	Error(lang string) string
	NewError() error
}

type userError struct {
	tt   *textTemplate
	data map[string]interface{}
}

func (ue userError) Error(lang string) string {
	return ue.tt.String(lang, ue.data)
}

func (ue userError) NewError() error {
	return errors.Errorf(ue.Error("en"))
}

func NewUserError(key string, data map[string]interface{}) UserError {
	tt, ok := text.Keys[key]
	if !ok {
		//undefined text template, create a definition
		//which consists of the key and a list of data value descriptions
		tt = &textTemplate{
			key:     key,
			Caption: map[string]string{},
			Data:    map[string]string{},
		}
		for n, v := range data {
			tt.Data[n] = fmt.Sprintf("type(%T).example(%s)", v, v)
		}

		text.Keys[key] = tt
		writeTextKeys()
	}
	return userError{
		tt:   tt,
		data: data,
	}
}

//structure of text translations file
type textFile struct {
	Keys map[string]*textTemplate `json:"keys"`
}

var text = textFile{
	Keys: map[string]*textTemplate{},
}

type textTemplate struct {
	Caption      map[string]string             `json:"caption" doc:"Caption template for each language code"`
	Data         map[string]string             `json:"data" doc:"Data values by name with a description generated by the code."`
	key          string                        //runtime copy of key
	compiledText map[string]*template.Template //runtime compiled templates
}

func (tt *textTemplate) String(lang string, data map[string]interface{}) (s string) {
	s = "---undefined---"
	defer func() {
		if s == "---undefined---" {
			s = fmt.Sprintf("%s(lang:%s)(data:%+v)", tt.key, lang, data)
		}
	}()

	if tt == nil {
		return
	}

	if tt.compiledText == nil {
		tt.compiledText = map[string]*template.Template{}
	}

	compiled, _ := tt.compiledText[lang]
	if compiled == nil {
		if text, ok := tt.Caption[lang]; ok {
			var err error
			compiled, err = template.New("text(" + tt.key + "," + lang + ")").Parse(text)
			if err != nil {
				log.Errorf("cannot compile template \"%s\": %+v", text, err)
			} else {
				tt.compiledText[lang] = compiled
			}
		}
	}

	if compiled != nil {
		b := bytes.NewBuffer(nil)
		if err := compiled.Execute(b, data); err != nil {
			log.Errorf("Cannot execute text template(key:%d,lang:%s): %+v", tt.key, lang, err)
		} else {
			s = b.String()
		}
	}
	return
}

const textTranslationsFilename = "./text_translations"

func init() {
	f, err := os.Open(textTranslationsFilename + ".json")
	if err == nil {
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&text); err != nil {
			panic(fmt.Sprintf("failed to decode text_translations from file %s: %+v", textTranslationsFilename, err))
		}
	}
}

func writeTextKeys() error {
	f, err := os.Create(textTranslationsFilename + ".new.json")
	if err == nil {
		defer f.Close()
		if err := json.NewEncoder(f).Encode(text); err != nil {
			panic(fmt.Sprintf("failed to encode text_translations to file %s: %+v", textTranslationsFilename, err))
		}
	}
	return nil
}