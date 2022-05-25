module github.com/jansemmelink/ussd2

go 1.17

replace (
	github.com/jansemmelink/utils2 => ../utils2
)

require github.com/jansemmelink/utils2 v0.0.0

require (
	github.com/google/uuid v1.3.0
	github.com/nats-io/nats.go v1.15.0
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20220315160706-3147a52a75dd // indirect
)
