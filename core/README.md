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
  *  wampire.core.router.sessions: Audit active router sessions
  *  wampire.core.broker.dump: Explore broker topics and subscribers
  *  wampire.core.dealer.dump: Explore dealer procedures and registrations

## Installation
Install library as package:
```
  go get -u github.com/marcosQuesada/wampire/core
```  

## Demo
 You can find WAMPire implementation on the [root of this repo](https://github.com/marcosQuesada/wampire), so clone it and take a look on the out of the box features, CLI and HTML clients.
 