Aegis
==========

A simple utility for deploying a Golang based Lambda with an API using 
AWS API Gateway's `ANY` method with a `{proxy+}` path to handle any request.

This results in a very easy solution for building serverless APIs with Go,
resulting in a single Lambda and a very minimal API Gateway.

### Getting Started

You'll need an AWS account of course. You'll also want to have your credentials
in your user's local directory where AWS CLI likes to keep them. If you already
use AWS CLI, then you won't have to do anything new. If not, getting your credentials
setup is probably easies by following AWS CLI instructions.

Install Aegis, then create an `aegis.yaml` file and configure your Lambda. 
Ensure your Go app uses the `HandleProxy` function from `"github.com/tmaiaroto/aegis/lambda"`.

You can reference the `example` directory from this repo to help you out.
Then in your Go project directory run:

```
aegis up
```

If everything is configured properly, this should upload your Lambda and setup an API for you.
The CLI output will return the URL to you, but you can of course also see this in your AWS console.

### Your Go Function

AWS API Gateway when used with Lambda Proxy requires a specific response format. It's not quite 
like your typical Lambda response. It's basically a JSON response with `statusCode`, `headers`,
and `body` keys.

So when building your Go Lambda, use the `HandleProxy` function from this package and return an 
`int`, `map`, `string` and `error`. The error will prompt a 500 response and will automatically 
fill in the body with the error message if no body is provided.

```
lambda.HandleProxy(func(ctx *lambda.Context, evt *lambda.Event) (int, map[string]string, string, error) {

	event, err := json.Marshal(evt)
	if err != nil {
		// If this body string is empty, the error message will be used.
		return 500, map[string]string{}, "", err
	}

	// This will simply return the event JSON that was passed in.
	// Note: It will contain more than just what was passed in the HTTP request.
	// API Gateway is configured to pass everything for your use. HTTP request type, request body,
	// path, querystring parameters, as well as API Gateway stage variables and other configuration info.
	return 200, map[string]string{}, string(event), nil

})
```

#### Getting More Fancy

Having one handler isn't really that fancy. Not only can the URL path be literally anything, but it also
supports `ANY` HTTP method. So you end up having to write a switch or bunch of if/elses or something 
less than pretty.

So to make this even nicer, Aegis has a handler that will act as a router of the sorts. It will let you 
register a function to handle incoming requests for any path and HTTP method. It isn't an HTTP router
per se, as it doesn't work with HTTP requests, but it reads very much the same way.

If you don't define a handler for a specific path or method, it will automatically return a 404 for you.

```
aegis := lambda.NewRouter()

aegis.GET("/", func(ctx *lambda.Context, evt *lambda.Event) (int, map[string]string, string, error) {
	// ... same old code here ...
})
```

#### Not Using Aegis Handler for your Lambda

What if you want to use another Lambda function? You can! Just keep mind it's a Lambda Proxy. This means
a specific JSON response is required in order for it to work with API Gateway. The response format is
as follows:

```
{
  "statusCode": "200",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{\"key1\":\"value1\",\"key2\":\"value2\",\"key3\":\"value3\"}"
}
```

NOTE: The `body` must be a string. If it's JSON then you'll need to escape it. API Gateway will return
this as JSON. Also note the `Content-Type` header.

If you want to use another Lambda function, you'll need to configure `aegis.yaml` appropriately.
It's under `lambda.sourceZip` config key, for example:

```
app:
  name: Example Using Another Lambda
  keepBuildFiles: true
lambda:
  sourceZip: YourFunction.zip
  functionName: your_function
api:
  name: Example Aegis API
  description: This API uses a Lambda function not built by Aegis
```

### About the Project

There's a growing list of serverless frameworks and utilities out there. Some are maturing quite
nicely given the technology is still changing. This project was built for a very specific purpose.
It focuses solely on running Go in a single AWS Lambda with API Gateway. If you need more than that, 
here's a list of serverless resources specifically for Go: https://github.com/SerifAndSemaphore/go-serverless-list

I suspect some of the existing serverless frameworks will support API Gateway with Lambda Proxy 
in the future. Again, this is a very new space and everyone's research is kinda in different directions,
but I imagine much of it will start to consolidate. I will try to keep this tool working with any AWS SDK 
changes or otherwise put a giant notice in this readme with an alternative solution.

So the other reason for this project is education. The code is simple and straight forward and
I've tried to leave hlepful comments. So you can certainly read through it and learn.

Special thanks to [@tj](https://github.com/tj) and [@mweagle](https://github.com/mweagle) for their
help with this. There were a few things that weren't clear along the way and they really took the 
time to help me out. Thanks to the other people and projects out there too.

Curious why the name Aegis? The "shield" metaphore comes from putting up an application that's 
protected from downtime as Lambda and API Gateway scale very well. It's also Athena's shield who 
is the goddess of crafts and wisdom. Honestly, I needed a name quick and while Aegis typically
associates with security, I thought this fit too. Call the binary whatever you want =)