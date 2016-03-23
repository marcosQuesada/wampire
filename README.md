[![Build Status](https://travis-ci.org/marcosQuesada/wampire.svg?branch=master)](https://travis-ci.org/marcosQuesada/wampire)

WAMPire implementations [WIP]
=============================
 Proof of concept on a basic WAMP library, it's heavily inspired on WAMP Turnpike implementation.
 Why do this? Just for the sake of coding :)
 
 Run the playground:
 
 * Start server on port 8000
```bash
go run main.go -port=8000
```
   Open your browser on: http://localhost:8000 and enjoy the chat demo

* CLI Client:
```bash
go run clients/cliClient/client.go -hostname=foo.server.bar -port=12000
``` 
   Enabled commands [WIP]
   * HELP
   * SUB topic
   * PUB topic
   
Where:
 * hostname: server host name (default hostname is localhost)
 * port: port where WAMP server is listen on (default port is 8888)
