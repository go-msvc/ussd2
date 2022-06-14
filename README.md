# DONE #
- examples:
    - runs on console/nats/rest (selected in config)
    - demo sos_credit menu item 1..3 with static and dynamic content
	- demo users to change profile settings
	- examples uses external sessions to run scalable instances
	- demonstrates dynamic prompts
	- demonstrates service calls before menus and after selections
	- text substitution from session data
	- language switch (examples/users)

# TODO #
- Do input validation and retry prompt
- Validation options in config, e.g. type is int with max configured to 10 for one prompt and 100 for another...
- Do Set with expressions
- Custom value types in sessions with validation built in and params e.g. min/max per prompt?
- Check error handling - what goes to user and what is logged
- Make service calls on NATS/HTTP/Mongo/SQL/...
    - (do support direct calls as well as MS to make standalone services)
- Define service test cases + service call stubs
- Service logic based on service results/session data -> which then change next items
- Dynamic enable/disable/show/hide of menu items
- Dynamic prompts
- Make default to use in-memory session with config to use external, which is needed for scaling
- Demo ext session in memory or in db like redis, mongo, sql...
- Back option and crumbs to continue going back
- Paging in USSD when data["maxl"] is defined (include encoding) - may be in caller but also needed in console... mmm?
- General service documentation
- Build service container for NATS or REST or both with ENV to select (container excl console) - see how we select server with ENV as it is nested config in JSON... and how do one set config in env for nested value without making it super complicated.
- Define and run service tests with stubs
- Determine test coverage
- Detect and report missing translations and list all translations, also able to use text ids with external translation config
- Demonstrate use of cache, db, http rest, other ms, queue, schedule, ...
- web app server (compared to HTTP REST, NATS, Console, ...)
- Load multiple item files, with inclusion or just everything in a folder...mmm?

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
