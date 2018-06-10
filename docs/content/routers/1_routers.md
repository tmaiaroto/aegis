+++
type = "router"
toc = "false"
+++

<h1 class="toc-ignore">Routers</h1>

First things first. The <a href="https://github.com/aws/aws-lambda-go" target="_blank">AWS Lambda Go package</a>
is used by Aegis. In fact, it's basically required for native Go support in Lambda. I don't think anyone wants
to write a competing package (that'd be silly). However, that package is designed to be lightweight. It does not
include a bunch of creature comforts. That's where Aegis routers come in handy.

Above all else, there is a default handler that can be used for any incoming Lambda event. This is called
the `DefaultHandler`. Aegis routers aim to make life easier without hiding all of the unknowns in life. You
can always handle incoming JSON event messages as maps.

However, there's some very helpful routers to help direct events to your handlers. You're probably looking
to use a framework because you don't want to write all that logic yourself, right?

One last quick note: All of the routers you see below can be used on the same Lambda function. Aegis does not
restrict your Lambda to handling just one type of event. The design of your functions is entirely up to you.