# DONE #
- examples/empty
    - runs on console/nats/rest
    - only first menu item
# TODO #
- Busy preventing item definition at run-time, and add dyn items to session...
	see Set and Final and later still have to do dyn prompt and menu...

- Do input validation and retry prompt
- Custom value types in sessions with validation built in and params e.g. min/max per prompt?
- Check error handling - what goes to user and what is logged
- Make service calls on NATS/HTTP/Mongo/SQL/...
    - (do support direct calls as well as MS to make standalone services)
- Define service test cases + service call stubs
- Configure external sessions (default in memory)
- Service logic based on service results/session data -> which then change next items
- Dynamic enable/disable of menu items
- Dynamic menus
- Dynamic prompts
- Back option
- Paging in USSD when data["maxl"] is defined (include encoding)
- General service documentation
- Build service container for NATS or REST or both with ENV to select (container excl console)
- Multi-language options + language preference, Text vs String



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
