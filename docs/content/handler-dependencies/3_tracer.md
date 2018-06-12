# Tracer

> How to use a different TraceStrategy

```go
AegisApp = aegis.New(handlers)
// For example, disabling tracing
AegisApp.Tracer = &aegis.NoTraceStrategy{}
AegisApp.Start()
```

> How Aegis Router traces internally

```go
// Records data in the tracing strategy. In this case, and by default, X-Ray annotations.
r.Tracer.Record("annotation",
    map[string]interface{}{
        "RequestPath": req.Path,
        "Method":      req.HTTPMethod,
    },
)

err = r.Tracer.Capture(ctx, "RouteHandler", func(ctx1 context.Context) error {
    // Makes the Tracer available inside your handler.
    // Capture() also applies the annotations from above in the case of XRayTraceStrategy
    d.Tracer = &r.Tracer
    return handler.handler(ctx1, d, &req, &res, params)
})
```

The `Tracer` dependency implements a `TraceStrategy` interface. By default, this is AWS X-Ray
defined as `XRayTraceStrategy`. There is also `NoTraceStrategy` available if you'd like to disable
tracing (or for unit tests).

However, you can add your own tracing strategy interfaces and use those instead if you wanted to
use something other than X-Ray.

Using a different strategy is as simple as setting the `Tracer` field on the Aegis interface.
In the examples, this is usually called `AegisApp`. Do keep in mind that this is a pointer.

Aegis' tracing works a little differently than AWS' X-Ray. Though you are always free to import
<span class="nowrap">`github.com/aws/aws-xray-sdk-go/xray`</span> to use yourself.

Aegis receives all events in one centralized function and then sends them out to the appropriate
router which in turn calls the matching handler. This is all internal and while there are a few
hooks available, no one wants to use those hooks to run tracing. It's simply not convenient nor
very conventional.

So, Aegis' `XRayTraceStrategy` has some fields, namely; `Annotation`, `NamespaceAnnotation`, `Metadata`,
and `Error`. These are maps (except `Error`, a singular error). So we are simply adding to them with
`Record()` and `Capture()` will loop through them and add on to the segment in AWS X-Ray. A different
interface implementing `TraceStrategy` may do something completely different. The `NoTraceStrategy`
that Aegis has will simply do nothing.

Every router should add some default annoatations and metadata automatically for you. Whether your
trace strategy handles that data is another story. So if you are using X-Ray, you don't need to add
the API Gateway request path or HTTP method. If you run Aegis with all the defaults, you're just going
to get traced API requests with meaningful annotations.

## Segments

`BeginSegment()` and `BeginSubsegment()` are somewhat universal concepts. AWS X-Ray has the concept
of segments and subsegments, but other tracing services may only deal with segments. Those segments
may even be implicit so you may find yourself not needing to implement these functions in a custom
strategy depending on the service you're using.

Regardless, all they do is help organize and segment your traces. In the case of X-Ray, Aegis will
automatically create an "Aegis" segment for you when running Aegis from the CLI or local web server.
When running in AWS Lambda, a segment is created automatically by Lambda for you.

This may not be the case if using a custom trace strategy. You may also want to create subsegments
to trace within your handlers. These are the functions to help you out.

`CloseSegment()` and `CloseSubsegment()` then help you define the end of those segments. In the
case of the default `XRayTraceStrategy`, closing a subsegment also sends data off to X-Ray. These
functions just call X-Ray's `Close()` and `CloseAndStream()`. They're aliases more or less.

## Recording Data

Regardless of the tracing service you ultimately use, the concept of `Record()` is pretty universal.
The idea is that you want to record certain bits of data with your tracing. AWS X-Ray calls these
"metadata," "annotations," "namespaced annotations," and a single "error" that can be sent along with
each trace.

The `XRayTraceStrategy` implementation of `Record()` will simply set some struct fields with the data
and nothing will actually be "sent" to AWS X-Ray until `Capture()` or `CaptureAsync()` is called.

The reason for this is because we aren't "streaming" a bunch of data to AWS X-Ray. We're "reporting"
on segments of our application. Your trace strategy may differ and there's no reason `Record()` couldn't
send "real-time" data along. It's just not how X-Ray works.

Another important thing to note here is that `Record()` can be called at any time without context.
In the case of Aegis, it's called before (and outside of) whatever `Capture()` is wrapping. So there
is no context. If we needed the context, we'd call `Record()` inside of `Capture()` which has context.
Or, the specific context we care about.

By the time `Tracer` is given to your handler, you do have the relevant context. So you could leverage
it to pull out whatever data you need.

The idea is that `Record()` shouldn't _require_ a context argument. While AWS X-Ray's internal methods
for recording annotations and metadata do take a context (for the segment), we don't know if other trace
services are going to use the context or not. So removing this requirement from the process keeps
`TraceStrategy` a lot more flexible as an interface.

## Tracing Functions

`Capture()` and `CaptureAsync()` (for goroutinues) are "wrappers." They wrap the functions you wish to trace.
In the case of AWS X-Ray, this captures things like execution time and more. Perhaps most important of all here
is that they can capture errors.

What exactly will be traced is going to vary by the trace service (or your own code if you roll your own).
In the case of AWS X-Ray, a lot is actually traced for you. How long HTTP connections took, how long it
took to marshal things, and more.

One of the biggest things you're going to see in X-Ray is how long your handlers took to execute and how long
your Lambda took to execute. This really starts to give you a good idea about performance and where your
bottlenecks are.

![X-Ray trace example](/aegis/img/xray-example-trace.jpg "X-Ray example trace")

An important thing to note about Lambda here is that X-Ray is also going to illustrate for you the difference
between a "cold" and "warm" start in Lambda. You'll notice in the screenshot above here that the Lambda took
over 800ms to run. Not terrible for API Gateway and all the OAuth work that was going on, but also not stellar.
Subsequent invocations were faster because the Lambda container was "warm" and some data retrieved from the IDP
service (such as the `DescribeUserPool` and `DescribeUserPoolClient` calls) were already cached. Things like the
well known ket set and so on.

As you learn to leverage caching methods and tune your service configurations, you'll be able to visually see
the performance change over time with X-Ray. It's an incredibly useful service.