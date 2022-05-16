package main

import (
	"context"
	"fmt"

	"bitbucket.org/vservices/ussd"
	"bitbucket.org/vservices/utils/errors"
	"bitbucket.org/vservices/utils/logger"
	_ "bitbucket.org/vservices/utils/ms/nats"
)

var log = logger.New()

func main() {
	logger.SetGlobalLevel(logger.LevelDebug)

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

	//temp item used for all menu items not yet implemented
	nyi := ussd.Final("nyi", "Not yet implemented")

	svc := ussd.NewService(
		ussd.Menu("sos_credit_menu", "SOS Crédit").
			With("SOS credit a un ami", promptDestNr, promptAmount, funcSosCreditAUnAmi).
			With("SOS Credit a TELMA", nyi).
			With("SOS offre a TELMA", nyi).
			With("Rembourser SOS", nyi).
			With("Aide", nyi))
	if err := svc.Run(); err != nil {
		panic(errors.Errorf("failed to run: %+v", err))
	}

	//todo:
	//- add example of input validation, e.g. amount or dest nr
	//- std phone nr validation for prompts - depending on network preferences
	//- retry prompt for invalid answer with a suitable message
	//- external calls for send SMS or HTTP or MS or DB
	//- plugin for service quota control (PCM count, or A-B count per day etc...)
	//- plugin for user preferences
	//- plugin for user details/authentication (e.g. when not used from ussd)
	//- console server
	//- HTTP REST server
}

func execSosCreditAUnAmi(ctx context.Context) ([]ussd.Item, error) {
	s := ctx.Value(ussd.CtxSession{}).(ussd.Session)
	msisdn := s.Get("msisdn").(string)
	amount := s.Get("amount").(string)
	destNr := s.Get("dest_nr").(string)

	//send sms to request credit from my friend
	if err := sendSms(
		msisdn,
		destNr,
		fmt.Sprintf("Please send me credit of %s!!!", amount),
	); err != nil {
		return nil, errors.Errorf("failed to send SMS to " + s.Get("dest_nr").(string))
	}
	return []ussd.Item{ussd.Final("", fmt.Sprintf("Request for %s sent to %s", amount, destNr))}, nil
}

func sendSms(from, to, text string) error {
	log.Debugf("SMS(%s->%s): \"%s\"", from, to, text)
	return nil //todo
}
