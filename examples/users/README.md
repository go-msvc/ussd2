# Users Example #
This service demonstrates a simple in-memory user profile service.
Run in on the console, with msisdn 27821234567 (already configured in config.json).

It demonstrates how:
- profile can be loaded before the main menu is displayed
- profile values can be updated and stored and show in the next session
- generic prompts are used to change both profile values
- language switch can be used

```
==========[ START      ]==========     (total=  0.000s                  now=08:07:36.938)
Profile
1. Name: Jan
2. Dob: 1973-11-18
3. Exit
reply > 1

Do you want to change your name?
1. Yes please
2. No, back to menu
reply > 2

Profile
1. Name: Jan
2. Dob: 1973-11-18
3. Exit
reply > 2

Do you want to change your dob?
1. Yes please
2. No, back to menu
reply > 1

Enter new value for your dob: 
reply > 1972-04-12

Profile
1. Name: Jan
2. Dob: 1972-04-12
3. Exit
reply > 1

Do you want to change your name?
1. Yes please
2. No, back to menu
reply > 1

Enter new value for your name: 
reply > Anne-Marie

Profile
1. Name: Anne-Marie
2. Dob: 1972-04-12
3. Exit
reply > 3

Goodbye.
----------[ RELEASE    ]----------     (total= 17.295s content=  0.001s now=08:07:54.234)


==========[ START      ]==========     (total=  0.000s                  now=08:07:55.224)
Profile
1. Name: Anne-Marie
2. Dob: 1972-04-12
3. Exit
reply > 3

Goodbye.
----------[ RELEASE    ]----------     (total=  1.881s content=  0.001s now=08:07:57.106)
```