---
title: "W3C Vehicle Information Service Specification ver 2 Reference Implementation Tutorial"
---
## W3C Vehicle Information Service Specification ver 2 Reference Implementation Tutorial

The W3C VISSv2 specification, soon to become a W3C standard, is developed at the [W3C/automotive github](https://github.com/w3c/automotive).
A reference implementation in the form of a server that exposes an interface according to the specification is developed on this repo.

Also found on this repo are implementations of other components that are needed to realize a communication tech stack that reaches from clients through the server and to the underlying vehicle system interface.

![VISSv2 communication tech stack](/automotive-viss2/images/WAII-tech-stack.jpg?width=40pc)

These software components (SwCs) can be categorized as follows:
* server
* clients
* data storage
* feeders
* tools

The tutorial describes each SwC category in a separate chapter.
It also contains a few Proof of concept (POC) examples, and information about installing,
building and running Golang based SwCs, a Docker containerization, and about some peripheral components.
