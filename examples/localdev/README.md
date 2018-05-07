# Local Dev Example

This example shows how to run Aegis locally while developing. It's often convenient to build your functions and APIs
and be able to test them locally before deploying. Otherwise, it can take some time to upload your function, so feedback 
isn't as immediate. CloudWatch can be combersome to work with as well, so it's nice to see a log in your terminal.

Other use cases include CI/CD tools. So there are a few optional to run your handler function(s) locally.

## Local HTTP API
The most common is a local HTTP server that transforms normal HTTP requests to `APIGatewayProxyRequest` compatible
events (`map[string]interface{}`) to flow through the handlers like normal. This is compatible with all the routing
and middleware. However, there will be quite a few headers missing (unless you similar fake ones via configuration), 
since it's not actually running through API Gateway.

To use this (assuming app is your Aegis interface), call: `app.StartServer()`

## Single Run CLI
Another option is a single run via the CLI. This allows your binary (or `go run main.go`) to take an `--event`
flag with a value that defines a path to a JSON file. Inside this file is the event message.

There's some other options here that can be configured via flags. This includes `--pretty` to enable pretty
printing of the response as well as `--nolog` to hide any logging done by the handler function(s). These are
boolean flags so you don't need to provide `--nolog=true` it can just be the flag itself.

For `APIGatewayProxyRequest` events, the body will be printed separate from the rest of the response (which includes 
headers and status code) when using pretty print. This is because otherwise, if it's a JSON response, the JSON string 
is escaped and not indented nicely.

To use this (assuming app is your Aegis interface), call: `app.StartSingle()`

### Configuring "Start" Listeners
The "local" or "stand alone" listeners can be configured a bit. See `StandAloneCfg` for more. Some options include 
the ability to set headers for the HTTP server. For the CLI there are options for hiding logging and pretty printing 
the responses. Again, CLI flags can be used to override these coded configurations in the case of `StartSingle()`.

For example, pretty printing may be enabled via coded configuration. Then when deployed through a CI/CD tool, the 
binary may be executed with it turned off along with hiding all logging. This way it can have a more compact output.

The single run option always exits and will do so with a 0/1 code. So it can be used with any CI/CD tool looking for 
successful tests before deploying.

Of course you'll likely want to set an environment variable or something to switch between `StartSingle()` and `Start()` 
at that point or some sort of different build since you can't have it exit when running in AWS Lambda.