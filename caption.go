package ussd

import (
	"strings"

	"bitbucket.org/vservices/utils/errors"
)

//caption stores text for language codes
//e.g. {"fr":"Bonyour", "af":"Goeie dag", "so":"Dumela", "en":"Hello"}
//the text can have double-moustache substitution from session data
//e.g. {"en":"Your current account balance is R{{balance}}."}
type CaptionDef map[string]string

func (def CaptionDef) Validate() error {
	if len(def) < 1 {
		return errors.Errorf("no langCode:text entries in this caption")
	}
	for langCode /*, text*/ := range def {
		if len(langCode) != 2 || strings.ToLower(langCode) != langCode {
			return errors.Errorf("language code \"%s\" is not 2-letter lowercase", langCode)
		}
		//text may be "" in some cases, have leading/trailing spaces, etc etc... so not validated
		//bad text will simply be returned to the user
	}
	return nil
}

func (def CaptionDef) Text(s Session) string {
	langCode := ""
	if s != nil {
		langCode, _ = s.Get("lang").(string)
	}

	//try to get current lang text (or default in langCode == "")
	text, ok := def[langCode]
	//if not defined and used a langCode, try default langCode == ""
	if !ok && langCode != "" {
		text, ok = def[""] //default
	}
	//if still not defined, and any lang is defined, use first randon one
	if !ok && len(def) > 0 {
		for _, text = range def {
			break
		}
	}

	//do substitution from session data
	//todo: should be precompiled template-type thingy... ideally something faster than full template...
	//...

	return text
}
