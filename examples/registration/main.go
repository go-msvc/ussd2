package main

import (
	"context"

	ussd "github.com/jansemmelink/ussd2"
	sessions "github.com/jansemmelink/ussd2/rest-sessions/client"
	"github.com/jansemmelink/utils2/errors"
)

func main() {
	if err := ussd.LoadItems("./items.json"); err != nil {
		panic(errors.Errorf("failed to load items.json: %+v", err))
	}

	//get main menu defined in the JSON file:
	var ok bool
	mainMenu, ok = ussd.ItemByID("main", nil)
	if !ok {
		panic("missing main")
	}

	//register user functions
	// start := ussd.Func("start", start)
	// ussd.Func("set_new_value", setNewValue)
	// ussd.Func("switch_lang", switchLang)

	//use external HTTP REST service for sessions
	s := sessions.New("http://localhost:8100")
	ussd.SetSessions(s)

	//create and run the USSD service:
	svc := ussd.NewService(start)
	if err := svc.Run(); err != nil {
		panic(errors.Errorf("failed to run: %+v", err))
	}
}

/*
 * MSISDN type is defined here as an example, but typically should be in a reusable package for all your services to work the same
 */
func init() {

}

var msisdnMaxLen = 13
var msisdnSubcriberNumberLen = 7 //e.g. 7 digits = "1234567" from msisdn = "27821234567"
var msisdnHomeCountryCode = "27" //e.g. "27" from 27821234567

type Msisdn string

func (m Msisdn) Validate() error {
	l := len(m)
	if l < msisdnSubcriberNumberLen || l > msisdnMaxLen {
		return errors.Errorf("length must be %d..%d digits")
	}
}

func Start(ctx context.Context) ([]ussd.Item, error) {
	//see if has msisdn (e.g. in web we do not have it)
	s := ctx.Value(ussd.CtxSession{}).(ussd.Session)
	m, ok := s.Get("msisdn").(string)
	if !ok {

	}

	return nil, nil
}
