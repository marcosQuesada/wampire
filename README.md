[![Build Status](https://travis-ci.org/marcosQuesada/wampire.svg?branch=master)](https://travis-ci.org/marcosQuesada/wampire)

WAMPire
=======
 WAMPire is an implementation of [WAMP v2 (Web Application Messaging Protocol)](http://wamp-proto.org/) router and clients written in golang, it's heavily inspired on [WAMP Turnpike implementation](https://github.com/jcelliott/turnpike), WAMP protocol is based on asynchronous messaging on **Publish/Subscribe** and **Remote Procedure Call** invocations, WAMPire is a proof of concept to explore the many posibilities of WAMP implementations.
 
 CLI client has full features support as a regular WAMP client, it's very useful to audit platform statys allowing router introspection.
 
## Milestones
- [x] WAMP basic profile router
- [x] CLI WAMP client
- [x] HTML WAMP client
- [ ] Challenge Response Authentication
- [ ] WAMP advanced profile 
- [ ] WAMP distributed architecture 

## Features 
 * WAMP basic profile
  * Publish/Subscribe 
  * RPC Call/Invocation/Yield/Result
 * RPC introspection tools
  *  wampire.core.router.sessions: Audit active router sessions
  *  wampire.core.broker.dump: Explore broker topics and subscribers
  *  wampire.core.dealer.dump: Explore dealer procedures and registrations

## Installation
Install library as package:
```
  go get -u github.com/marcosQuesada/wampire/core
```  

## Demo
  Clone it and take a look on the out of the box features, CLI and HTML clients.
 Install required dependencies and build:
 ```bash
 go build
 ```
 * Start server on port 8000
```bash
go run main.go -port=8000
```
   Open your browser on: http://localhost:8000 and enjoy the chat demo

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
   
 ###Register a procedure from the HTML client, and use it from CLI.
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
  Router introspection:
 ```bash
 call wampire.core.router.sessions
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
  
  Broker introspection:
  ```bash
 call wampire.core.broker.dump
 2016/03/27 22:45:34.533074 client.go:220: RESULT 5
 +--------+---------------------------+
 | TOPICS |           VALUE           |
 +--------+---------------------------+
 | topics | io.crossbar.demo.chat.ch1 |
 +--------+---------------------------+
 +---------------+--------------------------------------+
 | SUBSCRIPTIONS |                VALUE                 |
 +---------------+--------------------------------------+
 |            15 | a50f569d-0a6e-5927-53b9-14643df3a9cf |
 |            16 | 5bf52fcb-d07d-5b07-5f5e-f01ddc1a1546 |
 |            19 | 0fc2a3a4-b655-5252-4a67-8e7b65c69d0e |
 |            21 | b66f6c92-9e24-51ea-644d-31b551f301ab |
 +---------------+--------------------------------------+
 ```
 
  Dealer introspection:
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
