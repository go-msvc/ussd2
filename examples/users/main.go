package main

import (
	"context"

	ussd "github.com/jansemmelink/ussd2"
	_ "github.com/jansemmelink/ussd2/ms/console"
	sessions "github.com/jansemmelink/ussd2/rest-sessions/client"
	"github.com/jansemmelink/utils2/errors"
	"github.com/jansemmelink/utils2/logger"
	_ "github.com/jansemmelink/utils2/ms/nats"
	_ "github.com/jansemmelink/utils2/ms/rest"
)

var mainMenu ussd.Item

var log = logger.New() //.WithLevel(logger.LevelDebug)

func main() {
	//logger.SetGlobalLevel(logger.LevelDebug)
	if err := ussd.LoadItems("./items.json"); err != nil {
		panic(errors.Errorf("failed to load items.json: %+v", err))
	}

	//get main menu defined in the JSON file:
	var ok bool
	mainMenu, ok = ussd.ItemByID("main", nil)
	if !ok {
		panic("missing main")
	}
	start := ussd.Func("start", start)
	ussd.Func("set_new_value", setNewValue)

	//todo: before menu is displayed, ensure we got msisdn, needed to send SMS...
	//and possible load some user account details...

	//use external HTTP REST service for sessions
	s := sessions.New("http://localhost:8100")
	ussd.SetSessions(s)

	//create and run the USSD service:
	svc := ussd.NewService(start)
	if err := svc.Run(); err != nil {
		panic(errors.Errorf("failed to run: %+v", err))
	}
}

type Profile struct {
	Name string
	Dob  string
}

var profileByMsisdn = map[string]Profile{
	"27821234567": {Name: "Jan", Dob: "1973-11-18"},
}

func start(ctx context.Context) ([]ussd.Item, error) {
	s := ctx.Value(ussd.CtxSession{}).(ussd.Session)
	msisdn, ok := s.Get("msisdn").(string)
	if !ok || msisdn == "" {
		return []ussd.Item{
			ussd.FinalDef{Caption: ussd.CaptionDef{
				"en": "Invalid msisdn({{msisdn}})",
			}}.Item(s),
		}, nil
	}

	profile, ok := profileByMsisdn[msisdn]
	if !ok {
		return []ussd.Item{
			ussd.FinalDef{Caption: ussd.CaptionDef{
				"en": "Unknown msisdn({{msisdn}})",
			}}.Item(s),
		}, nil
	}

	log.Debugf("Loaded profile(%s):%+v", msisdn, profile)
	s.Set("name", profile.Name)
	s.Set("dob", profile.Dob)
	return []ussd.Item{mainMenu}, nil
}

func setNewValue(ctx context.Context) ([]ussd.Item, error) {
	s := ctx.Value(ussd.CtxSession{}).(ussd.Session)
	msisdn, _ := s.Get("msisdn").(string)
	fieldName, _ := s.Get("field_name").(string)
	oldValue, _ := s.Get("old_value").(string)
	newValue, _ := s.Get("new_value").(string)
	log.Debugf("Changing(%s.%s) from \"%s\" -> \"%s\"", msisdn, fieldName, oldValue, newValue)

	p, ok := profileByMsisdn[msisdn]
	if !ok {
		log.Debugf("MSISDN(%s) Profiles:%+v", msisdn, profileByMsisdn)
		return []ussd.Item{
			ussd.FinalDef{Caption: ussd.CaptionDef{
				"en": "Unknown msisdn({{msisdn}})",
			}}.Item(s),
		}, nil
	}

	switch fieldName {
	case "name":
		p.Name = newValue
		s.Set("name", p.Name)

	case "dob":
		p.Dob = newValue
		s.Set("dob", p.Dob)

	default:
		return []ussd.Item{
			ussd.FinalDef{Caption: ussd.CaptionDef{
				"en": "Unknown field name({{f_name}})",
			}}.Item(s),
		}, nil
	}

	profileByMsisdn[msisdn] = p
	return []ussd.Item{mainMenu}, nil
}
