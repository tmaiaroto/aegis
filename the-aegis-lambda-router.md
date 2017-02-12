When using API Gateway with an ANY route hooked to just one Lambda function, it became obvious that some sort of "router" was needed. Of course by the time the request payload hits the Lambda function we are beyond the normal HTTP request/response loop. The way in which Lambda is passing data to your Go application is via stdio so it's not exactly a router, but we can still think of it as one.

From the JSON passed via stdio, Aegis will construct a request that feels like a normal HTTP request when using Aegis' router. This makes it very natural to write your API. Here's an example:

```go
// Aegis Lambda router
// Handler
func hello(ctx *lambda.Context, evt *lambda.Event, res *lambda.ProxyResponse, params url.Values) {
    res.String(http.StatusOK, "Hello, World!")
}

func fallThrough(ctx *lambda.Context, evt *lambda.Event, res *lambda.ProxyResponse, params url.Values) {
    res.SetStatus(404)
}

func main() {
    // Aegis Lambda Router instance
    router := lambda.NewRouter(fallThrough)
    
    // Routes
    router.Handle("GET", "/", hello)
    
    // Listen (to stdio - hence no port)
    router.Listen()
}
```

Let's compare that with, the popular, [Echo](https://github.com/labstack/echo) router for example:

```go
// Echo router
// Handler
func hello(c echo.Context) error {
  return c.String(http.StatusOK, "Hello, World!")
}

func main() {
  // Echo instance
  e := echo.New()

  // Routes
  e.GET("/", hello)

  // Start server
  e.Logger.Fatal(e.Start(":1323"))
```

As you can see, it's very similar. In fact, most HTTP routers look like this and the goal with Aegis was to make writing serverless APIs feel very comfortable. Like many HTTP routers, Aegis' Lambda router also has support for middleware. It also features a "fall through" route that serves as a catch all. This is especially important because when using AWS API Gateway with an `ANY` request, literally _any_ request is technically valid. So your Lambda needs to handle them all or an error will be returned. Which is ok, but you may want a simple 404 instead.

So what does that mean for Lambda invocation? Well yes, that does mean you're invoking a Lambda for a 404. However, you can take advantage of API Gateway's cache feature and bypass any needless Lambda invocations if you decide to retire a route or have something that puts you in a situation like that.

The next sections will go over Aegis Lambda Router in more detail.

