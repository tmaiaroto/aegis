# Aegis Variables

> Access via Aegis interface

```go
barStr := AegisApp.GetVariable("foo")
```

> Access via HandlerDependencies

```go
func handler(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
    barStr := d.GetVariable("foo")
    return nil
}
```

> Access the map directly

```go
func handler(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
    if barStr, ok := d.Services.Variables["foo"]; ok {
        // ...
    }
    return nil
}
```

One of the most important helpers is the one to work with **_Aegis Variables_** which is just a conventional environment
variable. It could be an actual environment variable on the operating system, an AWS Lambda environment variable that
was configured, or an API Gateway stage variable.

So that means the helper function checks those locations in the following order:

 * Lambda environment variable
 * API Gateway stage variable
 * Actual operating system environment variable

The helper function is available on the `Aegis` interface or `HandlerDependencies`. A string value is always returned
because that's the only data type these variables can be.

You could also forgo the helper and access these variables on handler dependencies. It's considered a service.
So you'll find everything under `Services.Variables`.

Just note that the helper will check for an actual environment variable as a last resort. Accessing the `Services`
in handler dependencies will not include environment variables since this map is set by AWS Lambda and API Gateway
configurations.

Note that this works really well in conjunction with the `secret` CLI command. [See here for more info.](/aegis/cli/#secret)

