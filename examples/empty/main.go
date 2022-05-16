package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"

	"bitbucket.org/vservices/ussd"
	_ "bitbucket.org/vservices/ussd/ms/console"
	"bitbucket.org/vservices/utils/errors"
	"bitbucket.org/vservices/utils/logger"
	_ "bitbucket.org/vservices/utils/ms/nats"
	_ "bitbucket.org/vservices/utils/ms/rest"
)

var log = logger.New()

func main() {
	debugFlagPtr := flag.Bool("d", false, "Debug mode")
	flag.Parse()
	if *debugFlagPtr {
		logger.SetGlobalLevel(logger.LevelDebug)
	} else {
		logger.SetGlobalLevel(logger.LevelError)
	}

	//SosCreditAUnAmi:
	// Menu select 1
	// Entrer numero de tel. Destinataire :
	promptDestNr := ussd.Prompt("prompt_dest_nr", "Entrer numero de tel. Destinataire :", "dest_nr")
	// 111
	// Verifiez le numero de telephone SVP.
	// 341111111
	// Montant demande:
	promptAmount := ussd.Prompt("prompt_amount", "Montant demande:", "amount")
	// 200
	// Votre demande de recharge de 200 Ar a ete envoyee a 261341111111.
	funcSosCreditAUnAmi := ussd.Func("exec_sos_credit_a_un_ami", execSosCreditAUnAmi)

	//SOS Credit a TELMA:
	menuSelectAmount := ussd.Menu("select_amount", "Montant:").
		With("200", ussd.Set("amount", "200")).
		With("500", ussd.Set("amount", "500")).
		With("1000", ussd.Set("amount", "1000"))
		//With("back")	//todo: back must clear/replace []nextItems
	funcSosCreditATelma := ussd.Func("exec_sos_credit_a_telma", execSosCreditATelma)

	//register this func that is later called by id from dynamic menu
	/*funcSosApplyOffer :=*/
	ussd.Func("exec_sos_apply_offer", execSosApplyOffer)

	//temp item used for all menu items not yet implemented
	nyi := ussd.Final("nyi", "Not yet implemented")

	svc := ussd.NewService( //todo: ensure we got msisdn, needed to send SMS...
		ussd.Menu("sos_credit_menu", "SOS Crédit").
			With("SOS credit a un ami", promptDestNr, promptAmount, funcSosCreditAUnAmi).
			With("SOS credit a TELMA", menuSelectAmount, funcSosCreditATelma).
			With("SOS offre a TELMA", ussd.Func("get_offers", execSosGetOffers)).
			With("Rembourser SOS", nyi).
			With("Aide", nyi))
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
		return nil, errors.Errorf("failed to send SMS to " + s.Get("dest_nr").(string))
	}
	return []ussd.Item{ussd.DynFinal(s, fmt.Sprintf("Request for %s sent to %s", amount, destNr))}, nil
}

func execSosCreditATelma(ctx context.Context) ([]ussd.Item, error) {
	s := ctx.Value(ussd.CtxSession{}).(ussd.Session)
	//msisdn, _ := s.Get("msisdn").(string)
	amount, _ := s.Get("amount").(string)
	//todo: make service call to see if qualify and recharge account
	return []ussd.Item{ussd.DynFinal(s, fmt.Sprintf("Telma recharged your account with %s credit.", amount))}, nil
}

func execSosGetOffers(ctx context.Context) ([]ussd.Item, error) {
	s := ctx.Value(ussd.CtxSession{}).(ussd.Session)
	//msisdn, _ := s.Get("msisdn").(string)
	//todo: get offers applicable to this sub from external system
	//for now, hard coded:
	type Offer struct {
		Amount string
	}
	offers := []Offer{}
	offerAmount := 0
	for i := 0; i < rand.Intn(2); i++ {
		offerAmount += (rand.Intn(3) + 1) * 100
		offers = append(offers, Offer{Amount: fmt.Sprintf("%d", offerAmount)})
	}

	//if no offers - say sorry
	if len(offers) < 1 {
		return []ussd.Item{ussd.DynFinal(s, "Sorry, no offers available at present.")}, nil
	}

	//there are offers, present them as a menu then apply the selection
	def := ussd.DynMenuDef("Offers for You")
	for _, o := range offers {
		def = def.With(o.Amount, ussd.DynSet(s, "amount", o.Amount))
	}
	exec, _ := ussd.ItemByID("exec_sos_apply_offer", s)
	return []ussd.Item{def.Item(s), exec}, nil
}

func execSosApplyOffer(ctx context.Context) ([]ussd.Item, error) {
	s := ctx.Value(ussd.CtxSession{}).(ussd.Session)
	//msisdn, _ := s.Get("msisdn").(string)
	amount, _ := s.Get("amount").(string)
	//todo: make service call to see if qualify and recharge account
	return []ussd.Item{ussd.DynFinal(s, fmt.Sprintf("Telma recharged your account with %s credit.", amount))}, nil
}

func sendSms(from, to, text string) error {
	log.Debugf("SMS(%s->%s): \"%s\"", from, to, text)
	return nil //todo
}
