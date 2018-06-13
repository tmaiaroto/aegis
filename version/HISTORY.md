This file will include a bit of history/release notes.
Note that it is not a complete accounting of changes. It's a weak attempt at a more formal process.

## version 0.x

Before verison 1.x, aegis focused primarily on being a deploy tool for AWS Lambda and API Gateway.
It used a Node.js shim to get Go to work with Lambda pre-native Go support. It had a router and
some helpers. That's about it. The goal was educational as well as just a quick way to start a Go
project/API using Lambda.

## verison 1.x

This version marks a major milestone. Native Go support was made available for Lambda and starting
with this new major version, the Node.js shim was ditched. The decision was also made to start
expanding upon the simple convenient functions and handlers/routers. This is the beginning of
a more proper "framework." The intent is a lightweight framework and deploy tool.

Aegis is opinionated and prefers convention over configuration. It provides many helpers to
handle incoming messages and invents to Lambda functions. It also provides some relevant
AWS infrastructure/resource creation and management.

The goal is not to create a feature rich infrastructure management tool. Use something like
Terraform. The goal is not to create a framework for use with multiple clouds that supports
as many services, languages, and providers as possible. Use something like Serverless.

The goal is to do a few things well instead of trying to do everything. It's about providing
a lightweight set of helpers or framework to help build things faster. It's to be conventional
and flexible. 1.x will focus on adding more event router/handlers and helper functions.
Not every possible service will likely ever covered, the focus will be on the common.

## 1.14.1

- Added a bunch of tests (and more to come)
- Minor fixes exposed by tests

## 1.14.0

- Adjustments to `TraceStrategy` interface, making it an actual interface and making new
  `XRayTraceStrategy` struct to use by default and a new `NoTraceStrategy` to optionally
  use (and it's used for internal unit tests)
- Overhauled `TraceStrategy` interface methods as well, making it more generic
- Fix S3 Object router to also consider S3 object event name
- Fixed the configuration of S3 Object router's fall through handler
- Fixed S3 Object router's fall through handler match, using * with glob was uhhh, oops =)
  Changed the fall through key to _ across the board for consistency. It was a problem for
  anything using glob matching, but not an issue elsewhere.

## 1.12.1

- Fixed (changed) `Tasker` handlers to use `map[string]interface{}` instead of pointer
  to map since that's silly becuase maps are reference types

## 1.12.0

- Moved local HTTP server to Aegis interface; it will eventually handle more than just
  HTTP requests (which mimicks APIGatewayProxyRequest) - this creates a very small breaking
  change, the old `router.Gateway()` is simply replaced by `app.StartServer()` where app
  is an instance of Aegis interface
- Moved test cases accordingly, many of the functions to convert requests/responses
  were simply moved
- Added `StartSingle()` which takes an `--event` flag when running the binary for the
  event and returns a pretty printed result (optional) or error message to the CLI
  (also can flag `--pretty` for this and `--nolog` to hide any log output from the app)

## 1.11.1

- Added test cases
- Fixed Aegis handler Filters to pass `evt` by value not pointer
- Also altered function signature of `After` filter to include both the `interface{}` and
  the `error` that normally is returned from the handler
- Add `HTMLError()` helper function
- Fixed `XMLError()` helper function

## 1.11.0

- Added SES integration; configure rules in `aegis.yaml` and handle with `SESRouter` handlers
- Fixed bug when handling with nil Router, it will now return an error explaining that no
  handleres have been set to handle the event (this can occur if a router was registered,
  but not added to aegis.Handlers{}, leading to some confusion)

## 1.10.1

- Added the ability to use standard middleware (Go's http.Handler interface)
- Added `update` CLI command for updating Lambda function code only (faster than full deploy)
- Fix import path case issue for logrus

## 1.10.0

- Add ability to work with AWS Secret Manager from CLI `aegis secret` command
- Can now set Lambda environment variables and API Gateway stage variables in `aegis.yaml` 
  by referring to values in AWS Secret Manager in the format of: `<secretName.key>`
- `aegis.yaml` API Gateway stage variables now support case sensitive keys by defining
  values as maps in the YAML since Viper lowercases keys
- Helper functions added to retrieve "Aegis Variables" which is a convention around
  the priority of variables to use (Lambda environment, API GW stage, os env)
- Added `Custom` (`map[string]interface{}`) to `HandlerDependencies` for custom needs

## 1.9.3

- Added Router level middleware (it runs first before the individual route's)

## 1.9.2

- Added a `HostedLogoutURL` field to the `CognitoAppClient` struct. This will help
  handlers automatically redirect users to Cognito hosted logout page.
- Added a Cognito example
- Added a basic "hello world" example
- Fixed/updated `aegis init` boilerplate (that hello world example)

## 1.9.1

 - `lambdaHandler` was added to the Handlers interface for backward compatibility.

## 1.9.0

Unfortunately there are a few breaking changes with this release.
Technically speaking that might dictate a 2.x release, but since this is all "beta"
I'd like to bend semantic versioning rules here.

This is a major improvement that gets the framework out of an architectural bind.
The handler function signature is changing and while there were options to avoid it,
they involved asking the developer to use wrappers/closures or to always use type 
assertions. This leads to potential panics. It also involved using Context for DI,
which is frowned upon (for good reason). Lambda events in Go already use interfaces
and there are still events unhandled by the Lambda Go package so not everything comes
into the handlers as a struct. It's not a good idea to move even farther away from 
type safety. Not for a framework.

As much as I hate breaking changes and debated the options extensively...I hate 
sticking others (or my future self) with a potential runtime explosion even more.
I believe in conventions and reasonable defaults -- not magic. Not obscurity.

- Service dependency injection into handlers has been added
- ***Breaking change*** (small): The signature of _all_ event handlers is now:
  `handler(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIProxyRequest ...)`
  It puts a `HandlerDependencies` struct as the second arg and shifts down the event map/structs.
  So it's a relatively small change to all handlers, the arg can be safely ignored if
  you don't need it, but it's still a breaking change.
- *Breaking change* (minor): `Cognito` was moved to a general `Services` field on Aegis.
  `configurations` was moved under `Services` as well. This supports the DI efforts.
- DI now makes possible a generic Cognito JWT middleware that you can use.
  You can find it under `cognito_helpers.go`. It relies on an `access_token` cookie
  and helps reduce some boilerplate code for you.
- Filters have been added to Aegis interface eventHandle() function allowing
  "application-wide" behavior changes or interceptions. It is possible to change
  the Lambda event map for everything at once. It is possible to "re-configure"
  services based on incoming events, etc. Likely a rarely used feature, it adds
  some flexibility for situations that have no other option.
- "Tracer" (TracingStrategy) has been added to the injected handler dependencies.
  This makes it much easier for end users to trace AND add to the existing tracing
  done by the framework (add annotations, metadata, etc.). When it comes to XRay,
  this means working directly with the same subsegment.

## 1.8.0

- Added Cognito app client interface/helper (verify tokens, etc.)
- Adjusted deploy command to add Cognito access to aegis lambda role
- Added cookie helpers using Go's http package by creating fake http.Request
  so reading cookies from API Gateway requests is now easy
- Added an Aegis interface that helps manage various dependencies and services 
  to capitalize on Lambda container re-use
- *Breaking change* (minor): RPC() now invokes Lambdas without tracing support.
  So context is not required as the frist argument. This allows an RPC call to be
  made outside of a Lambda event handler. An additional RPC() method has been added
  to the new Aegis interface. This uses the context set on Aegis for tracing (which
  can be overridden). This function also does not require context to be passed.
  This helps to eliminte passing dependencies just for things like tracing.
- AWSClientTracer alias has been removed

## 1.7.0

- RPC handler receiver function changed to take evt value instead of pointer 
  as a pointer would be redundant for map, better practice this way
- being support for Cognito trigger events which also take and return a map value

## 1.6.0

- S3 bucket notification triggers and router handler
- begin restructuring/organization of functions for deploy command (Deployer interface)
