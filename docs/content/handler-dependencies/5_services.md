# Services

> From the Cognito example Aegis app

```go
AegisApp.ConfigureService("cognito", func(ctx context.Context, evt map[string]interface{}) interface{} {
    // Automatically get host. You don't need to do this - you typically would have a known domain name to use.
    host := ""
    stage := "prod"
    if headers, ok := evt["headers"]; ok {
        headersMap := headers.(map[string]interface{})
        if hostValue, ok := headersMap["Host"]; ok {
            host = hostValue.(string)
        }
    }
    // Automatically get API stage
    if requestContext, ok := evt["requestContext"]; ok {
        requestContextMap := requestContext.(map[string]interface{})
        if stageValue, ok := requestContextMap["stage"]; ok {
            stage = stageValue.(string)
        }
    }

    return &aegis.CognitoAppClientConfig{
        // The following three you'll need to fill in or use secrets.
        // See README for more info, but set secrets using aegis CLI and update aegis.yaml as needed.
        Region:   "us-east-1",
        PoolID:   AegisApp.GetVariable("PoolID"),
        ClientID: AegisApp.GetVariable("ClientID"),
        // This is just automatic for the example, you would likely replace this too
        RedirectURI:       "https://" + host + "/" + stage + "/callback",
        LogoutRedirectURI: "https://" + host + "/" + stage + "/",
    }
})
```

Handler dependencies contain `Services` too. These are primarily AWS services like AWS Cognito for example.
They are a bit different than a simple field on the `Aegis` interface because they are "services" that can
carry configurations which often need to be dynamic.

However, services are really just dependencies as well. Your normal dependencies, set on the `Aegis`
interface can not be configured dynamically. You set the dependency on the interface and that's it.

Using `ConfigureService()` is the important function here. The first argument for this function is the name
of the service you wish to configure. The second argument is a function. That function is given the context
and the event from Lambda. However, this configuration is applied before your event handler is called.

<aside class="not-warning">
Take care when configuring services since you have access to the Lambda event. You could accidentally
alter the event in a way your handler is not expecting.
</aside>

## Filters

There are hooks or "filters" available to use here for the handler as well. The Aegis interface has
a field called `Filters` for various filters, the ones most helpful to services is likely the
<span class="nowrap">`Filters.Handler.BeforeServices`</span> filter.

## Custom Services
3rd party services can also be added and configured. They are available under the `Custom` field.

TBD - not fully implemented