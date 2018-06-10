# RPC Router

> Route and handle by name

```go
func main() {
    rpcRouter := aegis.NewRPCRouter()
    rpcRouter.Handle("lookup", handleLookup)

    app := aegis.New(aegis.Handlers{
        // Again, your one function can handle different event types
        // Router: router,
        RPCRouter: rpcRouter,
    })
    app.Start()
}

func handleLookup(ctx context.Context, d *aegis.HandlerDependencies, evt map[string]interface{}) (map[string]interface{}, error) {
    // Some GeoIP look would happen here and some data would be returned
    return map[string]string{"city": "Somewhereville"}, nil
}
```

> Example GeoIP lookup from another Lambda

```go
rpcPayload := map[string]interface{}{
    "_rpcName":  "lookup",
    // or req.IP() if req is an APIGatewayProxyRequest
    "ipAddress": req.RequestContext.Identity.SourceIP,
}
resp, rpcErr := aegis.RPC("aegis_geoip", rpcPayload)
```

This interface allows "RPC" (remote procedure calls) to be handled. In Aegis' world, that is to say a Lambda invoking
another Lambda. So the immediate question you should have is, "how do you know it's an invocation from another Lambda?"
Great question! We don't. Not really anyway. Plus, who's to say it has to be another Lambda invoking it? Technically,
an RPC could come from any program that has access.

The way that this router matches is, like `Tasker`, through conventions in the JSON message payload itself. In this case,
instead of looking for a `_taskName` key, it'll be <span class="nowrap">`_rpcName`</span> instead.

<aside class="note-warning">
<i class="fas fa-exclamation-triangle"></i> Could your scheduled task that invokes your Lambda via CloudWatch rule trigger
event also use an RPC handler? Yes, it could, but the idea is to be organized and leverage these conventions so you don't
end up with unpredictable behavior.
</aside>

Like `Tasker`, `RPCRouter` also has an optional fall through handler that you can pass to `NewRPCRouter()` when setting
up the router. The handler function signature is very close to the `Tasker` handler functions as well.

Unlike most other handlers, this one requires you to return more than just an error. A `map[string]interface{}` must be
returned as well. What good would the RPC be without something coming back?

Technically speaking, Aegis returns data on your `Router` handlers as well. It's just done for you automatically because
working with API Gateway responses is predictable. The response struct is what gets returned and Aegis does that for you.
However, in the case of an RPC, it's impossible to know exactly what you are going to return other than a map, because
we know that Lambda works with JSON messages.

Aegis has a top level help function, `RPC()`, that makes it a bit easier to invoke another Lambda. If you are looking
to invoke one Lambda from another, this should just work for you without any extra effort as the default IAM role on your
Lambdas through Aegis will permit Lambda invocation.

If your needs are more complex than a simple call with a function name and map payload, you will need to use the AWS SDK
to invoke the Lambda instead. Or if you didn't use Aegis to deploy your Lambdas (which would have set up an IAM role),
and you have other special considerations - again, feel free to do your thing. Aegis' helpers are not the only way to
go about things.

<aside class="note-info">
<i class="fas fa-info-circle"></i> Don't forget to return a map and an error (or nil) when handling RPCs.
</aside>