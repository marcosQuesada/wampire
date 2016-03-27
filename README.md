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
   
   