# DONE #
- examples/empty
    - runs on console/nats/rest (selected in config)
    - demo sos_credit menu item 1..3 with static and dynamic content
# TODO #
	- get external sessions to demonstrate scaling with NATS and REST

- complete and test dynamic prompts and language code in session
- Do input validation and retry prompt
- Do Set with expressions
- Captions with translation and substitution
- Custom value types in sessions with validation built in and params e.g. min/max per prompt?
- Check error handling - what goes to user and what is logged
- Make service calls on NATS/HTTP/Mongo/SQL/...
    - (do support direct calls as well as MS to make standalone services)
- Define service test cases + service call stubs
- Configure external sessions (default in memory)
- Service logic based on service results/session data -> which then change next items
- Dynamic enable/disable/show/hide of menu items
- Dynamic prompts
- Back option and crumbs to continue going back
- Paging in USSD when data["maxl"] is defined (include encoding)
- General service documentation
- Build service container for NATS or REST or both with ENV to select (container excl console)
- Multi-language options + language preference, Text vs String
- Determine test coverage



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
