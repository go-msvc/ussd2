module bitbucket.org/vservices/ussd

go 1.18

replace (
	bitbucket.org/vservices/utils => ../utils.js2
)

require bitbucket.org/vservices/utils v0.0.0

require (
	github.com/google/uuid v1.3.0
	github.com/nats-io/nats.go v1.15.0
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20220315160706-3147a52a75dd // indirect
)
