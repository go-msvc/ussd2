package main

import (
	"bitbucket.org/vservices/ussd"
	"bitbucket.org/vservices/utils/errors"
	"bitbucket.org/vservices/utils/logger"
	_ "bitbucket.org/vservices/utils/ms/nats"
)

func main() {
	logger.SetGlobalLevel(logger.LevelDebug)
	//SosCreditAUnAmi:
	// Menu select 1
	// Entrer numero de tel. Destinataire :
	//promptDestNr := ussd.Prompt("Entrer numero de tel. Destinataire :")
	// 111
	// Verifiez le numero de telephone SVP.
	// 341111111
	// Montant demande:
	//promptAmount := ussd.Prompt("Montant demande:")
	// 200
	// Votre demande de recharge de 200 Ar a ete envoyee a 261341111111.
	//execSosCreditAUnAmi := ussd.Func()

	nyi := ussd.NewFinal("nyi", "Not yet implemented")
	svc := ussd.NewService(
		ussd.Menu("sos_credit_menu", "SOS Crédit").
			With("SOS credit a un ami", nyi). //, promptDestNr, promptAmount, execSosCreditAUnAmi).
			With("SOS Credit a TELMA", nyi).
			With("SOS offre a TELMA", nyi).
			With("Rembourser SOS", nyi).
			With("Aide", nyi))
	if err := svc.Run(); err != nil {
		panic(errors.Errorf("failed to run: %+v", err))
	}
}
