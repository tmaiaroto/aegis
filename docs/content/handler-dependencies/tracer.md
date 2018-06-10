# Tracer

The `Tracer` dependency implements a `TracingStrategy` interface. By default, this is AWS X-Ray
defined as `XRayTraceStrategy`. There is also `NoTraceStrategy` available if you'd like to disable
tracing (or for unit tests).

However, you can add your own tracing strategy interfaces and use those instead if you wanted to
use something other than X-Ray.

Aegis' tracing works a little differently than AWS' X-Ray. Though you are always free to import
<span class="nowrap">`github.com/aws/aws-xray-sdk-go/xray`</span> to use yourself.

Aegis receives all events in one centralized function and then sends them out to the appropriate
router which in turn calls the matching handler. This is all internal and while there are a few
hooks available, no one wants to use those hooks to run tracing. It's simply not convenient nor
very conventional.

> How Aegis Router traces internally

```go
// Trace/capture the handler (in XRay by default) automatically
r.Tracer.Annotations = map[string]interface{}{
    "RequestPath": req.Path,
    "Method": req.HTTPMethod,
}

err = r.Tracer.Capture(ctx, "RouteHandler", func(ctx1 context.Context) error {
    r.Tracer.AddAnnotations(ctx1)
    r.Tracer.AddMetadata(ctx1)
    r.Tracer.AddErrors(ctx1)

    // Tracer is available to you in your handler
    d.Tracer = &r.Tracer
    return handler.handler(ctx1, d, &req, &res, params)
}
```

So, Aegis' `XRayTraceStrategy` has some fields, namely; `Annotations`, `Metadata`, and `Errors`.
These are maps and slices. So you simply add to them your annotations, metadata, and errors. Aegis
then calls `TracingStrategy` methods; `AddAnnotations()`, `AddMetadata()`, and `AddErrors()` which
will loop through those fields to add them.

In other words, you are instructing the `Tracer` what to add at the time the event is handled.
Any other tracing you wish to do is up to you.

In some cases, routers will add some default annoatations and metadata automatically for you.
So you don't need to add the API Gateway request path or HTTP method for example. If you run Aegis
with all the defaults, you're just going to get traced API requests with meaningful annotations.

However, there are a lot of similarties between Aegis' `Tracer` interface and AWS X-Ray's package.
That's because Aegis is designed to use X-Ray. So many of the function names and concepts align.
Though you should still be able to implement your own tracing strategies. You simply may not use
all of the methods on the interface.

