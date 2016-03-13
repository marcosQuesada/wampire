[![Build Status](https://travis-ci.org/marcosQuesada/wampire.svg?branch=master)](https://travis-ci.org/marcosQuesada/wampire)

WAMP protocol implementation [WIP]
==================================
 Proof of concept on a basic WAMP library, it's heavily inspired on WAMP Turnpike implementation.
 Why do this? Just for the sake of coding :)
 
## Milestones
 * WAMP basic server (supports HELLO/WELCOME, SUBSCRIBE/SUBSCRIBED, PUBLISH/PUBLISHED)
 * WAMP RPC dealer (supports CALL / RESULT)

### WIP
 Inside tools folder there's a basic server&client implementation, you can use it that way:
 
 * Server running on port 12000
```bash
go run main.go -port=12000
```
* Client:
```bash
go run client.go -hostname=foo.server.bar -port=12000
```
Where:
 * hostname: server host name (default hostname is localhost)
 * port: port where WAMP server is listen on (default port is 8888)
