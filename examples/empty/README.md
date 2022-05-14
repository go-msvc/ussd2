# Example #
This is an empty USSD service, just to test that it is valid USSD service
that can be used by the `ms-vservices-ussd-router`.

It supports an operation `ussd` which has the standard vservices ussd.Request/Response
interface. But then based on the request type, it will call the new operation for either
`start`, `continue` or `abort`.

So there are in total the following operations:
- `ussd` (for any request)
- `start` to start a new session
- `continue` to continue an existing session with user input.
- `abort` to end a session (usually when user killed it on the phone)
