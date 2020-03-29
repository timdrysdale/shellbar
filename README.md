# Shellbar


![alt text][logo]

![alt text][status]
![alt text][coverage]

Shellbar relays websocket streams for secure point-to-point communications

## Why?

Remote laboratory experiments are often located behind firewalls without the prospect of installing a conventional jumpserver to aid administration. While there are a number of *nix tools that could do this job on a one-off basis with just a bit of script-glue (various combinations of ```socat``` and ```websocat```), there is no existing solution that

+ is well-suited to management of large numbers remote experiments
+ can handle fully interactive shells without some gremlins, such as fatal connection loss on Ctrl+C
+ supports multiple simultaneous independent client connections to the same service 
+ allows all clients and servers to simultaneously connect to the relay on the same port yet keep traffic separate

## Features

+ binary and text websockets
+ separate treatment for servers and clients
+ servers requires custom client to multiplex the individual client streams (in development at [timdrysdale/sa](https://github.com/timdrysdale/sa))
+ clients can use any websocket client they like
+ multiple 1:1 streams
    + organised by topics
	+ topics set by routing
	+ multiple clients can connect to the same topic for 1:1 connection with server
	+ when multiple clients connect to the same topic, they can not see each other's traffic
	+ only servers can see the traffic from clients
+ statistics are recorded, and available via
    + websocket API in JSON format
	+ human-readable webpage with 5sec update at ```<host>/stats```
	
## What's with the name?
It's a variation on a websocket relay called [timdrysdale/crossbar](https://github.com/timdrysdale/crossbar). ```Shellbar``` is mainly intended for relaying shell commands, e.g. via ```ssh``` for interactive sessions and remote administration e.g. via [ansible](https://www.ansible.com). The two projects address different parts of the practable™ ecosystem, so it make senses to keep their development separate.

## Getting started

Binaries will be available in the future, but for now it is straightforward to compile.

It's quick and easy to create a working ```go``` system, [following the advice here](https://golang.org/doc/install).

Get the code, test, and install
```
$ go get github.com/timdrysdale/crossbar
$ cd `go list -f '{{.Dir}}' github.com/timdrysdale/shellbar`
$ go test ./cmd
$ go install
```
To run the relay
```
$export SHELLBAR_LISTEN=:8888
$ shellbar
```

Navigate to the stats page, e.g. ```http://localhost:8888/stats```

You won't see any connections, yet.

You can connect using any of the other tools in the ```practable``` ecosystem, e.g. [timdrysdale/vw](https://github.com/vw/timdrysdale) or if you already have the useful [websocat](https://github.com/vi/websocat) installed, then

```
websocat --text ws://127.0.0.1:8888/expt - 
```
If you type some messages, you will see the contents of the row change to reflect that you have sent messages.

## Applications

### Relaying shell access

```ssh``` is a server-speaks-first protocol, which has influenced the design toward using a custom websocket client at the server, which works with the relay to the keep the data streams separate for each client.

## Usage

connect with a suitable ```shellbar``` client to the server routing, ```<yourhost>/serve/some/routing```. Connect with clients to ```<yourhost>/some/routing```. Each client will have a separate 1:1 stream with the server. The ```shellbar``` client at the server end is required to make a new connection to the service every time a new client connects to the ```shellbar``` relay, so that server-speaks-first protocols like ```ssh``` work.

## Support / Contributions

Get in touch! My efforts are going into system development at present so support for alpha-stage users is by personal arrangement/consultancy at present. If you wish to contribute, or have development resource available to do so, I'd be more than delighted to discuss with you the possibilities.

[status]: https://img.shields.io/badge/alpha-in%20development-red "Alpha status; in development"
[coverage]: https://img.shields.io/badge/coverage-54%25-orange "Test coverage 56%"
[logo]: ./img/logo.png "Shellbar logo - shells connected in a network to a letter S"
