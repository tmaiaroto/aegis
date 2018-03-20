# Aegis

[![License Apache 2](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://github.com/tmaiaroto/aegis/blob/master/LICENSE) [![godoc aegis](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/tmaiaroto/aegis) [![Build Status](https://travis-ci.org/tmaiaroto/aegis.svg?branch=master)](https://travis-ci.org/tmaiaroto/aegis) [![Go Report Card](https://goreportcard.com/badge/github.com/tmaiaroto/aegis)](https://goreportcard.com/report/github.com/tmaiaroto/aegis)

A simple utility for deploying a Golang based Lambda with an API using   
AWS API Gateway's `ANY` method with a `{proxy+}` path to handle any request.

This results in a very easy solution for building serverless APIs with Go,  
resulting in a single Lambda and a very minimal API Gateway.

### Getting Started

You'll need an AWS account of course. You'll also want to have your credentials  
in your user's local directory where AWS CLI likes to keep them. If you already  
use AWS CLI, then you won't have to do anything new. If not, getting your credentials  
setup is probably easiest by following AWS CLI instructions. Note that you can   
also pass keys via the CLI or by setting environment variables.

Install Aegis, then create an `aegis.yaml` file and configure your Lambda.

AWS Lambda for Go works a little differently than when working with Node.js functions.
Each handler function will be passed a given event type instead of a JavaScript object.
This means, you must write handlers that use certain structs.
See: https://github.com/aws/aws-lambda-go/tree/master/events

The most common handler is likely to handle requests from AWS API Gateway.
So you can reference the `example` directory from this repo to help you out there. Or you can  
copy some example files to get you started with using API Gateway and Lambda with Go:

```
aegis init
```

Then in your Go project directory run:

```
aegis deploy
```

If everything is configured properly, this should upload your Lambda and setup an API for you.  
The CLI output will return the URL to you, but you can of course also see this in your AWS console.

Aegis has a handler that will act as a router of the sorts. It will let you register a function to 
handle incoming requests for any path and HTTP method. It isn't an HTTP router per se, as it doesn't 
work with HTTP requests/responses, but it reads very much the same way.

The router also supports middleware.

```go

import aegis "github.com/tmaiaroto/aegis/framework"

func main() {
    router := aegis.NewRouter(fallThrough)
    router.Handle("GET", "/", root)
    router.Listen()
}

func fallThrough(ctx context.Context, evt *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
    res.StatusCode = 404
    return nil
}

func root(ctx context.Context, evt *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
    res.Body = "body for root path"
    res.Headers = map[string]string{"Content-Type": "text/plain"}
    return nil
}
```

Note that structs in Aegis' framework package (APIGatewayProxyRequest, APIGatewayProxyResponse, etc.) are
simply references to the underlying AWS Go Lambda package's structs. However, there is some additional
functionality composed on to them with the router.

#### Logging

Go's normal logging will work and end up in CloudWatch. Additionally, logrus is available under `lambda.Log`  
and a hook has been added for CloudWatch. Additional hooks can be added for other centralized logging solutions.

#### Testing Locally

Sometimes it's handy to test your Lambda function before deploying to AWS. Aegis allows you to do so if you  
are using its router. While a `router.Listen()` is used for Lambda \(via stdio\), a `router.Gateway()` is   
used for starting a local web server \(you can configure the port by setting `router.GatewayPort` it's  
`:9999` by default\).

When writing your app using the router and making this gateway, you can simply use `go run main.go`   
for example and test your Lambda without even building your app let alone deploying to AWS.

This makes it very convenient to test while you develop, but keep in mind that not all data  
normally found in a Lambda Event and Context are available to you.

Also keep in mind that there is no Lambda for IAM roles to be assigned to. So your local AWS credentials   
will need to be valid for any AWS services you want to use.

### About the Project

There's a growing list of serverless frameworks and utilities out there. Some are maturing quite  
nicely given the technology is still changing. This project was built for a very specific purpose.  
It focuses solely on running Go in a single AWS Lambda with API Gateway. If you need more than that,   
here's a list of serverless resources specifically for Go: [https://github.com/SerifAndSemaphore/go-serverless-list](https://github.com/SerifAndSemaphore/go-serverless-list)

I suspect some of the existing serverless frameworks will support API Gateway with Lambda Proxy   
in the future. Again, this is a very new space and everyone's research is kinda in different directions,  
but I imagine much of it will start to consolidate. I will try to keep this tool working with any AWS SDK   
changes or otherwise put a giant notice in this readme with an alternative solution.

So the other reason for this project is education. The code is simple and straight forward and  
I've tried to leave hlepful comments. So you can certainly read through it and learn.

Special thanks to [@tj](https://github.com/tj) and [@mweagle](https://github.com/mweagle) for their  
help with this. There were a few things that weren't clear along the way and they really took the   
time to help me out. Thanks to the other people and projects out there too.

