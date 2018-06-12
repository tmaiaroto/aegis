+++
type = "handler-dependencies"
toc = "false"
+++

<h1 class="toc-ignore">Handler Dependencies</h1>

Handlers need a variety of services and dependencies in order to make life easier. It's a very big
part of why one reaches for a framework. The idea of "dependency injection" is the foundation of many
frameworks and Aegis isn't much different.

Really, routing events to handlers is the foundation of Aegis, but you can't abstract that away without
being able to pass along functionality required by each handler.

Aegis has a few core dependencies that even its internal handler (a handler before your handler)
uses and passes along for you to use as well. Also, you can also pass along your own dependencies
if you need to.

