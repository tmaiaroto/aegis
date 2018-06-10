# About Aegis
### An AWS Serverless Framework for Golang

Aegis is both a simple deploy tool and framework. It's primary goal is to help you develop microservices in the AWS cloud quickly and easily. They are mutually exclusive tools.

The Aegis CLI is not intended to be an infrastructure management tool. It will never be as feature rich as tools like Terraform. It's goal is to assist in the development of microservices - not the maintenance of infrastructure.

Likewise the framework is rather lightweight as well. It may never have helpers and features for every AWS product under the sun. It provides a conventional framework to help you build serverless microservices faster. It removes a lot of boilerplate.

# Getting Started

If you use Go and AWS already, you likely can skip the first step. If that's the case, the rest should take you about 5 minutes
to get your first app up and running.

### Step 1.

For starters, you're going to need <a href="https://golang.org/" target="_blank">Go</a> on your development machine. You'll also need
an <a href="https://aws.amazon.com/" target="_blank">Amazon Web Services</a> account. Obvious enough, right?

Then, make sure you have your <a href="https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/setup-credentials.html" target="_blank">
Amazon credentials setup for development.</a> You should also be able to set environment variables when you use the `aegis` command
line tool, but do you really want to do that each time? It gets annoying to type them out. Set the environment variables in a more
permanent fashion or use an `~/.aws/credentials` file.

It also wouldn't hurt for you to have <a href="https://aws.amazon.com/cli/" target="_blank">AWS CLI tools.</a> Though they shouldn't
technically be necessary.

### Step 2.

Then, you'll want to get Aegis. Go makes this easy, just run: <span class="nowrap">`go get github.com/tmaiaroto/aegis`</span>

This should provide you with the `aegis` binary. If you can't run it from your command line, then you may need to add your Go
bin path to your `PATH` (and however Windows, Powershell or whatever does it). This is probably a good time to also note that
the Aegis CLI tool mostly assumes you're using Mac OS or Linux. It has not extensively been tested on Windows. Please feel free
to submit some issues should there be any.

### Step 3.

Go to a fresh directory where you want to create a new Go package and Aegis project, ie.
<span class="nowrap">`~/go/src/github.com/your-user/your-project`</span>

From there, run the following command: `aegis init`

This will provide you with a boilerplate `main.go` file to get you started as well as an `aegis.yaml` file. You'll likely want
to open the YAML file and edit the names of things. Name your function, API, etc. Then you'll work with the Go file.

### Step 4.

To deploy, run: `aegis deploy`

It will create an IAM role, the Lambda function, API Gateway, and any other resources needed. You'll see some output in your terminal
with information about what's going on, where to access the API, etc.

Congratulations, you should have a serverless application deployed now. Head on over to the <a href="/aegis/routers/">Routers section</a>
to learn about how to handle various events.

# Goals & Philosophies

Knowing what Aegis _is_ and knowing what _does_ and _why_ are different things. Don't use any tool or framework based on what it is, you
want to use one based on what it does and why it does it. More to the point, you want to use a framework if its goals and philosphies are
in line with your own project's needs.

Clearly, if you're still reading, your projects uses Go or you are considering using Go. Ok, but what about some of these other
considerations?

## DevSecOps

One of the goals of this project is to provide consideration for "development security operations" or DevSecOps. This is kind of an
extension of normal DevOps in that it also brings security into the fold.

A wonderful example here is with regard to managing sensitive credentials. Aegis CLI works with AWS Secrets Manager in order to help
you keep sensitive credentials, such as database passwords, protected. Aegis' deploy tool will then insert those values into Lambda
environment variables or API Gateway stage variables. This prevents accidental publishing of credentials through version control,
CI/CD tools, logging, and so on.

Aegis also uses XRay by default to help you gain visibility into your stack, though you can switch it out for something else too.

## Convention over Configuration

It's a recurring theme in many solid frameworks. Complex configurations will slow down development and increase maintenance. Where it
makes sense, having sensible defaults and conventions is smart. For example, accepting any request through API Gateway to then handle
through a code based router is much nicer than configuring each route in API Gateway and then each handler in code. Simply ignore
(or handle through a "fall through" or "catch all") the types of requests that aren't needed.

The handling of various events types through a "router" or pattern matching is yet another convention or pattern that you'll pick
up and use.

You can't completely avoid configuration of course. There is an `aegis.yaml` file afterall and AWS requires a lot of things to be
configured. However, where it makes sense, you'll see the convetion over configuration theme througout the framework.

## Serverless Patterns

There are several approaches when building serverless functions and "microservices." You can take the approach of logically
separating your Lambdas to handle all events for a certain service (like Domain Driven Design). Another approach is to handle
just one event for a service per Lambda resulting in many different functions.

Aegis allows you to architect your application in any way you like. It doesn't force you into creating one Lambda per API Gateway
route or per event handled. You can if you like, but you don't need to. You're free to come up with your own conventions.

## Vendor Lock-In

Yes, there is a little bit of that going on here. The decision to use AWS was made. A line in the sand was drawn. Though you
are free to cross that line in some cases. Just because your application runs on AWS Lambda, doesn't mean you can't use services
from other cloud providers.

You could decide to use bug tracking and logging from some other service if you like. You can use a serverless database like
CosmosDB on Microsoft Azure if you like. Nothing prevents you from this.

However, Aegis the deploy tool and framework will never allow you to use the same code to deploy either an AWS Lambda or an
Azure Function. The framework's goal is not to abstract so you can "run anywhere." The reason for this is because the framework
would be far more complex and bloated if that were the case. It would also become a lot more limited too because some features
exist in one cloud provider but then not another.

So you're going to need to use Amazon here. There's just no way around it. _If_ Aegis were to ever support another cloud provider
then there would be more of a "port" to another framework and deploy tool instead.

This also means that all core services are Amazon services. You won't find a service to work with any other provider in the core
framework. However, you might certainly find other Go packages to import into your application code to use other services outside
of Amazon.

Services can be injected as dependencies into handlers, so it's pretty straight forward to configure and use something from
another provider in your handlers with some sanity on configuration and code re-use.

# Additional Reading

Outside of this documentation here, you can find a variety of articles that illustrate both how to use Aegis as well as some
example use cases for AWS Lambda with Aegis.

[A Cognito Protected API in Minutes](https://serifandsemaphore.io/a-cognito-protected-serverless-api-with-golang-in-minutes-a054c9f50cf3)    
Shows how to quickly deploy an Aegis API that uses a Cognito to secure it.

[Quickly Deploy a GeoIP Microservice](https://serifandsemaphore.io/aws-lambda-geoip-golang-microservice-with-aegis-91bae736c1b2)    
Shows you how to quickly deploy a microservice that returns the requesting client's geolocation based on their IP.