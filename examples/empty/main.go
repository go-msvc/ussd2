package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"

	ussd "github.com/jansemmelink/ussd2"
	_ "github.com/jansemmelink/ussd2/ms/console"
	sessions "github.com/jansemmelink/ussd2/rest-sessions/client"
	"github.com/jansemmelink/utils2/errors"
	"github.com/jansemmelink/utils2/logger"
	_ "github.com/jansemmelink/utils2/ms/nats"
	_ "github.com/jansemmelink/utils2/ms/rest"
)

var log = logger.New()

var (
	startMenu ussd.Item
)

func main() {
	debugFlagPtr := flag.Bool("d", false, "Debug mode")
	flag.Parse()
	if *debugFlagPtr {
		logger.SetGlobalLevel(logger.LevelDebug)
	} else {
		logger.SetGlobalLevel(logger.LevelError)
	}

	//register functions written in go
	//so they can be referenced from other items and in the json file
	ussd.Func("func_sos_credit_a_un_ami", execSosCreditAUnAmi)
	ussd.Func("func_sos_credit_a_telma", execSosCreditATelma)
	ussd.Func("func_get_offers", execSosGetOffers)
	ussd.Func("exec_sos_apply_offer", execSosApplyOffer)

	//other types of items can be defined in code and referenced from JSON too
	//but this could also be defined in JSON file with:
	//	"nyi":{"final":{"caption":{"fr":"Not yet implemented"}}}
	ussd.Final("nyi", ussd.CaptionDef{"fr": "Not yet implemented item({{item_id}})"})

	if err := ussd.LoadItems("./items.json"); err != nil {
		panic(errors.Errorf("failed to load items.json: %+v", err))
	}

	//todo: resolve all next item references not yet defined

	//get start item defined in the JSON file:
	var ok bool
	startMenu, ok = ussd.ItemByID("start", nil)
	if !ok {
		panic(errors.Errorf("start item not found"))
	}

	//todo: before menu is displayed, ensure we got msisdn, needed to send SMS...
	//and possible load some user account details...

	//use external HTTP REST service for sessions
	s := sessions.New("http://localhost:8100")
	ussd.SetSessions(s)

	//create and run the USSD service:
	svc := ussd.NewService(startMenu)
	if err := svc.Run(); err != nil {
		panic(errors.Errorf("failed to run: %+v", err))
	}
}

func execSosCreditAUnAmi(ctx context.Context) ([]ussd.Item, error) {
	s := ctx.Value(ussd.CtxSession{}).(ussd.Session)
	msisdn, _ := s.Get("msisdn").(string)
	destNr, _ := s.Get("dest_nr").(string)
	amount, _ := s.Get("amount").(string)

	//send sms to request credit from my friend
	if err := sendSms(
		msisdn,
		destNr,
		fmt.Sprintf("Please send me credit of %s!!!", amount),
	); err != nil {
		return nil, errors.Errorf("failed to send SMS to " + destNr)
	}
	return []ussd.Item{
		ussd.FinalDef{
			Caption: ussd.CaptionDef{
				"fr": fmt.Sprintf("Request for %s sent to %s", amount, destNr),
			},
		}.Item(s),
	}, nil
}

func execSosCreditATelma(ctx context.Context) ([]ussd.Item, error) {
	s := ctx.Value(ussd.CtxSession{}).(ussd.Session)
	//msisdn, _ := s.Get("msisdn").(string)
	amount, _ := s.Get("amount").(string)
	//todo: make service call to see if qualify and recharge account
	return []ussd.Item{
		ussd.FinalDef{
			Caption: ussd.CaptionDef{
				"fr": fmt.Sprintf("Telma recharged your account with %s credit.", amount),
			},
		}.Item(s),
	}, nil
}

func execSosGetOffers(ctx context.Context) ([]ussd.Item, error) {
	s := ctx.Value(ussd.CtxSession{}).(ussd.Session)
	//msisdn, _ := s.Get("msisdn").(string)
	//todo: get offers applicable to this sub from external system
	//for now, hard coded:
	type Offer struct {
		Name   string
		Amount int
	}
	offers := []Offer{}
	offerAmount := 0
	nrOffers := rand.Intn(5)
	for i := 0; i < nrOffers; i++ {
		offerAmount += (i + 1) * 100
		offers = append(offers, Offer{
			Name:   fmt.Sprintf("%d Ar", offerAmount),
			Amount: offerAmount,
		})
	}

	//if no offers - say sorry
	if len(offers) < 1 {
		return []ussd.Item{
			ussd.FinalDef{
				Caption: ussd.CaptionDef{
					"fr": "Sorry, no offers available at present.",
				},
			}.Item(s),
		}, nil
	}

	//there are offers, present them as a menu then apply the selection
	exec, _ := ussd.ItemByID("exec_sos_apply_offer", s)
	menuDef := ussd.DynMenuDef(ussd.CaptionDef{"fr": "Offers for You"})
	for _, o := range offers {
		menuDef = menuDef.With(ussd.CaptionDef{"fr": o.Name},
			ussd.SetDef{Name: "offer_name", Value: o.Name}.Item(s),
			ussd.SetDef{Name: "amount", Value: o.Amount}.Item(s),
			exec,
		)
	}
	menuDef = menuDef.With(ussd.CaptionDef{"fr": "Back"},
		startMenu,
	)
	//log.Debugf("Defined offers menu: %+v", menuDef)
	return []ussd.Item{menuDef.Item(s)}, nil
}

func execSosApplyOffer(ctx context.Context) ([]ussd.Item, error) {
	s := ctx.Value(ussd.CtxSession{}).(ussd.Session)

	//msisdn, _ := s.Get("msisdn").(string)
	offerName, _ := s.Get("offer_name").(string)
	amount, _ := s.Get("amount").(int)
	log.Infof("Add credit of %d Ar...", amount)
	//todo: make service call to see if qualify and recharge account
	return []ussd.Item{
		ussd.FinalDef{
			Caption: ussd.CaptionDef{
				"fr": fmt.Sprintf("Telma recharged your account with %s.", offerName),
			},
		}.Item(s),
	}, nil
}

func sendSms(from, to, text string) error {
	log.Debugf("SMS(%s->%s): \"%s\"", from, to, text)
	return nil //todo
}
