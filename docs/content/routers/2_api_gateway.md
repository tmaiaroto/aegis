# API Gateway Router

> This should look familiar if you've built an HTTP RESTful API in Go before

```go
package main

import aegis "github.com/tmaiaroto/aegis/framework"

func main() {
    // Handle an APIGatewayProxyRequest event with a URL request path Router
    router := aegis.NewRouter(fallThrough)
    router.Handle("GET", "/", handleRoot)

    // Register the handler
    app := aegis.New(aegis.Handlers{
        Router: router,
    })
    // A blocking call that listens for events
    app.Start()
}

// fallThrough handles any path that couldn't be matched to another handler
func fallThrough(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
    res.StatusCode = 404
    return nil
}

// handleRoot is handling GET "/" in this case
func handleRoot(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
    res.JSON(200, map[string]interface{}{"event": req})
    return nil
}
```

Perhaps the most common event handler is for API Gateway requests. When you think about a router for a web
application, typically handling HTTP requests comes to mind. So the Aegis' router for this is simply `Router`.

While the concept of serverless applications and "event handling" brings about far more opportunties to "route"
events, this interface takes the less the descriptive name purely due to familiarity.

The `Router` works with an `ANY` method on a wildcard path in API Gateway. Instead of defining paths and methods
to associate with different Lambda functions, Aegis prefers a simpler approach. Though you aren't bound to this,
it certainly makes things a lot easier. However, keep in mind that your Lambda function will be triggered by
API Gateway on literally any request.

Amazon's official <a href="https://github.com/aws/aws-lambda-go" target="_blank">AWS Lambda Go</a> package will
marshal the incoming JSON event message to an API Gateway Proxy Request struct of some sort. Aegis does exactly
the same thing, in fact, it aliases Amazon's struct. The AWS Lambda Go package handles quite a number of different
type of events by converting them into native Go structs. However, you can always handle a plain old map too.

Aegis adds helper functions on to several AWS Lambda Go structs. The goal is to make it feel more familiar for anyone
who has ever written an API in Go before. Also take a look at the `res` variable in the example code. See how you
can easily set `res.StatusCode` and return data with the `res.JSON()` helper function?

Note that every single handled event in Aegis has a return value, which is an `error`. You'll most often simply
return `nil`, but if you do return an actual error, it will be returned in the response in this case. Other routers
may have no where to return the error of course, but a [tracer](/handler-dependencies/#tracer) might handle it
(for example, AWS X-Ray will handle it).

You could also return your own error with appropriate status code with body content. For example, you could return
a JSON body response that contains the error with a status code of 500. There's even a helper function for this as
well: `res.JSONError(500, e)` That will return the text of the error in a JSON message.

## APIGatewayProxyRequest

This struct will contain a good bit of information from API Gateway. Most important of all, it will include HTTP
request headers, body, and querystring parameters.

There are a bunch of helpers that you can read about in the [HTTP (Proxy) Helpers section](/helpers/#http-proxy-helpers).
Examples include; `IP()` which returns the IP address of the client. `GetForm` which returns form-data from the request.
`Cookies()` which returns the cookies from the request. These are convenient functions that AWS' core Go Lambda package
does not provide, but Aegis' aliased version does.

## APIGatewayProxyResponse

This struct is responsible for the response from handled HTTP requests. Ultimately everything in AWS Lambda comes
in as JSON and goes out as JSON. API Gateway will transform the response as needed, but Lambda deals with JSON for
its messaging format.

Therefore, `APIGatewayProxyResponse` is unmarshaled by the AWS Lambda Go package but, you get to work with a nice
struct that can have functions composed on to it. Again, helpers for this response include things like setting status
codes, body content, headers, and more.

Like `APIGatewayProxyRequest`, the response struct is also an alias of the AWS Lambda Go package.

## Fall Through Handler

When creating a new `Router` you can define a "fall through" or "catch all" handler as seen in the example code snippet.
This will handle any request not matched by your router. A common handler in this case is simple one that does nothing,
but sets a `StatusCode` of 404.

<aside class="note-info">
<i class="fas fa-info-circle"></i> Some routers in Aegis have optional fall through handlers. The API Gateway Router's fall through is not optional.
</aside>

## Middleware

Aegis' router also supports middleware. Standard Go http library middleware at that.