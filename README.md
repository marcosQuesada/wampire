[![Build Status](https://travis-ci.org/marcosQuesada/wampire.svg?branch=master)](https://travis-ci.org/marcosQuesada/wampire)

WAMPire
=======
 WAMPire is an implementation of [WAMP v2 (Web Application Messaging Protocol)](http://wamp-proto.org/) router and clients written in golang, it's heavily inspired on [WAMP Turnpike implementation](https://github.com/jcelliott/turnpike), WAMP protocol is based on asynchronous messaging on **Publish/Subscribe** and **Remote Procedure Call** invocations, WAMPire is a proof of concept to explore the many posibilities of WAMP implementations.
 
 CLI client has full features support as a regular WAMP client, it's very useful to audit platform statys allowing router introspection.
 
## Milestones
- [x] WAMP basic profile router
- [x] CLI WAMP client
- [x] HTML WAMP client
- [x] Routed RPC
- [x] Session Meta Procedures
- [X] Session Meta Events
- [ ] Challenge Response Authentication
- [ ] WAMP advanced profile 
- [ ] WAMP distributed architecture 

## Features 
 * WAMP basic profile
  * Publish/Subscribe 
  * RPC Call/Invocation/Yield/Result
 * RPC introspection tools
  * wampire.session.list: List all active sessions
  * wampire.session.count : Count all active sessions
  * wampire.session.get : Introspect a Session
  * wampire.core.broker.dump: Explore broker topics and subscribers
  * wampire.core.dealer.dump: Explore dealer procedures and registrations

## Installation
Install library as package:
```
  go get -u github.com/marcosQuesada/wampire/core
```  

## Demo
  Clone it and take a look on the out of the box features, CLI and HTML clients.
 Install required dependencies and build:
 ```bash
go get -t ./...
 ```
 * Start server on port 8000
```bash
go run main.go -port=8000
```
   Open your browser on: http://localhost:8000 and enjoy the chat demo, web client buit with [Autobahn.js](http://autobahn.ws/)

 * CLI Client:
```bash
go run clients/cliClient/client.go -hostname=foo.server.bar -port=12000
``` 
Where:
 * hostname: server host name (default hostname is localhost)
 * port: port where WAMP server is listen on (default port is 8888)

 CLI enabled commands
   * HELP
   * LIST: list all registered procedures
   * SUB topic
   * PUB topic message
   * CALL uri arguments
   * EXIT (or Ctl+C)

## Session Meta Events
  Server publishes session updates on an special topic: 
   ```bash
    wampire.session.meta.events
    ```
  Session updates:
    *wampire.session.on_join
    *wampire.session.on_leave
  Broker updates:  
    *wampire.subscription.on_create
    *wampire.subscription.on_subscribe"
    *wampire.subscription.on_unsubscribe"
    *wampire.subscription.on_delete
  Dealer Updates:
    *wampire.registration.on_register
    *wampire.registration.on_unregister
    
## Examples
 Let's interact between CLI and HTML clients.
 ###Subscribe both clients to the same topic and talk between them
  From browser client, select desired channel (Ch1), making a subscribe to the topic io.crossbar.demo.ch1,
  now open CLI client:
  ```bash
     sub io.crossbar.demo.chat.ch1
     2016/03/27 22:35:17.660630 client.go:253: Subscribed Details: %!s(core.ID=16) 
     pub io.crossbar.demo.chat.ch1 hello
     2016/03/27 22:35:29.716898 client.go:247: published Details: %!s(core.ID=2) 
  ```
   Message has been received from web client, make your reply and it will be received from CLI.
   
### Register a procedure from the HTML client, and use it from CLI.
 Open HTML client and register io.crossbar.demo.mul2 procedure, that procedure just 
 makes the product of two arguments, so click register button, and then, from a terminal:
 
```bash
   call io.crossbar.demo.mul2 2 2
   2016/03/27 22:35:48.111166 client.go:220: RESULT 3
   +------+----------------+
   | LIST |     VALUE      |
   +------+----------------+
   | list | %!s(float64=4) |
   +------+----------------+
```

### RPC introspection tools:
  Wampire registered procedures:
```bash
   call wampire.core.list
   2016/03/28 00:37:12.978673 client.go:224: RESULT 2
   +------+--------------------------+
   | LIST |          VALUE           |
   +------+--------------------------+
   | list | wampire.core.echo        |
   | list | wampire.session.list     |
   | list | wampire.session.count    |
   | list | wampire.session.get      |
   | list | wampire.core.broker.dump |
   | list | wampire.core.dealer.dump |
   | list | wampire.core.help        |
   | list | wampire.core.list        |
   +------+--------------------------+
```
  Router introspection, list current sessions:
```bash
 call wampire.core.session.list
 2016/03/27 22:45:25.085312 client.go:220: RESULT 4
 +------+--------------------------------------+
 | LIST |                VALUE                 |
 +------+--------------------------------------+
 | list | 2b3888d5-e830-5280-5c9f-44c38bba5ccc |
 | list | 5bf52fcb-d07d-5b07-5f5e-f01ddc1a1546 |
 | list | a50f569d-0a6e-5927-53b9-14643df3a9cf |
 | list | 0fc2a3a4-b655-5252-4a67-8e7b65c69d0e |
 | list | b66f6c92-9e24-51ea-644d-31b551f301ab |
 | list | d28237b0-e46f-56e4-76aa-81cc9f253372 |
 +------+--------------------------------------+
```
  
 Router total active session:  
```bash
  call wampire.session.count
  
 +------+----------------+
 | LIST |     VALUE      |
 +------+----------------+
 | list | %!s(float64=3) |
 +------+----------------+
```
 
 Session introspection:
```bash
 call wampire.session.get cf50900f-7989-54fa-42ab-a92b803a0abf
 2016/03/28 00:46:27.238950 client.go:224: RESULT 1
 +---------------+---------------------------+
 | SUBSCRIPTIONS |           VALUE           |
 +---------------+---------------------------+
 |            21 | io.crossbar.demo.chat.ch1 |
 +---------------+---------------------------+
 initTs: 2016-03-28T00:46:05.907272244+02:00
 +---------------+-----------------------+
 | REGISTRATIONS |         VALUE         |
 +---------------+-----------------------+
 |            22 | io.crossbar.demo.mul2 |
 +---------------+-----------------------+
```  
 
 Broker introspection, list all subscribers:

```bash
 call wamp.subscription.list_subscribers
 2016/03/28 01:12:44.759733 client.go:224: RESULT 5
 +---------------+--------------------------------------+
 | SUBSCRIPTIONS |                VALUE                 |
 +---------------+--------------------------------------+
 |            33 | a6fda6b3-0374-5658-7945-bdbb339631e7 |
 |            34 | a6fda6b3-0374-5658-7945-bdbb339631e7 |
 |            27 | 9ed64ba9-7aa6-57bd-4257-543958f4950e |
 |            29 | a6fda6b3-0374-5658-7945-bdbb339631e7 |
 |            30 | e274d51a-c9ad-59db-476f-097ca58b10ab |
 |            31 | e274d51a-c9ad-59db-476f-097ca58b10ab |
 |            32 | e274d51a-c9ad-59db-476f-097ca58b10ab |
 +---------------+--------------------------------------+
```
  
 Count all subscribers
```bash
 call wampire.subscription.count_subscribers
 2016/03/28 01:13:09.280094 client.go:224: RESULT 7
 +------+----------------+
 | LIST |     VALUE      |
 +------+----------------+
 | list | %!s(float64=7) |
 +------+----------------+
```
 
 List all topics:
```bash
 call wampire.subscription.list_topics
 2016/03/28 01:12:56.471340 client.go:224: RESULT 6
 +--------+---------------------------+
 | TOPICS |           VALUE           |
 +--------+---------------------------+
 | topics | io.crossbar.demo.chat.ch2 |
 | topics | io.crossbar.demo.chat.ch1 |
 | topics | io.crossbar.demo.chat.ch3 |
 | topics | fooTopic                  |
 +--------+---------------------------+
```
 
 List all subscribed sessions by topic:
```bash
 call wamp.subscription.list_topic_subscribers io.crossbar.demo.chat.ch1
 2016/03/28 01:13:39.663491 client.go:224: RESULT 8
 +------+--------------------------------------+
 | LIST |                VALUE                 |
 +------+--------------------------------------+
 | list | a6fda6b3-0374-5658-7945-bdbb339631e7 |
 | list | e274d51a-c9ad-59db-476f-097ca58b10ab |
 | list | 9ed64ba9-7aa6-57bd-4257-543958f4950e |
 +------+--------------------------------------+
```

  Dealer introspection, procedures and registrations:
```bash
 call wampire.core.dealer.dump
 2016/03/27 22:45:41.861125 client.go:220: RESULT 6
 +---------------+--------------------------------------+
 | REGISTRATIONS |                VALUE                 |
 +---------------+--------------------------------------+
 |             8 | internal                             |
 |            10 | internal                             |
 |            12 | internal                             |
 |            17 | a50f569d-0a6e-5927-53b9-14643df3a9cf |
 |             2 | internal                             |
 |             4 | internal                             |
 |             6 | internal                             |
 +---------------+--------------------------------------+
 +------------------------------+-----------------+
 |       SESSION HANDLERS       |      VALUE      |
 +------------------------------+-----------------+
 | wampire.core.dealer.dump     | %!s(float64=12) |
 | wampire.core.echo            | %!s(float64=6)  |
 | wampire.core.help            | %!s(float64=2)  |
 | wampire.core.list            | %!s(float64=4)  |
 | wampire.core.router.sessions | %!s(float64=8)  |
 | io.crossbar.demo.mul2        | %!s(float64=17) |
 | wampire.core.broker.dump     | %!s(float64=10) |
 +------------------------------+-----------------+
```
