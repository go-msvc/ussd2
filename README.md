# DONE #
- examples/empty
    - runs on console/nats/rest
    - only first menu item
# TODO #
- Busy preventing item definition at run-time, and add dyn items to sessions...
	see Set and Final and Menu, still have to do prompt
	Need to register tyd definitions so they can load from file
	And need to make func Xxx() take session always so that when load from file,
	  s == nil and it implies static, but when used at runtime, implies dynamic
	  and make it a session function if possible so that cannot be called without...
	  may be session.Set() and session.Menu().With()....Item() and session.Prompt() etc

- Do input validation and retry prompt
- Do Set with expressions
- Text with translation and substitution
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
