# Architecture & Goals

The goal of Aegis is to allow a developer, using Go, to build serverless APIs very quickly. One should be able to write some Go code and have an API up within minutes. The deploy process is fast. Convention over configuration is very much in mind with the decisions made around the tool. You deploy using the tool's CLI and while there is a config file to work with, you won't really be touching it all too often.

There's only a few main things to keep in mind with Aegis. 

1. Go is the only supported language
2. AWS is the only service provider supported \(Lambda & API Gateway\)
3. A single Go application can handle multiple requests
4. Any request is accepted through each API and passed to the Go application

Like anything, there are trade offs here. Aegis is very opinionated and because of that it does a few things really well. If you're not interested in writing a Lambda with Go, then you won't really benefit from this tool. Check out the [Serverless Framework](https://serverless.com/) instead. If you want to use Go, but separate out all of your Lambda functions and have a bit more of a "framework" with broader support, versioning and more check out [Apex](http://apex.run/) instead. If you want to use Go, but also want a bit more features with regard to other AWS services, discovery, etc. then check out [Go Sparta](http://gosparta.io/).

Aegis is "lighter weight" and has less features than all of the frameworks mentioned above. Yet all of these tools are assuming to a degree. The challenge with serverless right now is the disparity between service providers, their features, and their internals. AWS Lambda, Google Cloud Functions, Microsoft Azure Functions, Iron.io Functions, and Apache OpenWhisk \(and I'm probably forgetting some\) are all different from one another. In my opinion -- building a serverless framework or tool that attempts to normalize them or use them interchangeable is perhaps ambitious, but very unwise decision at this point in time. The field is far too new and volatile.

So you need to choose for now. You'll want to understand the strengths and weaknesses of each of these services in order to do that. However, you also want to choose one that's comfortable for you.

Again, my opinion is that AWS currently provides the most options for building a serverless RESTful API. A big factor in this is API Gateway and it's ability to proxy any HTTP request to a Lambda function \(among many other features including caching\).

#### So how does Aegis work?

It uses API Gateway's `ANY` request to handle any HTTP request method and it uses a `{proxy+}` path resource to handle any URL path. This means the API only has one configured route. This removes a huge amount of boilerplate configuration and keeps things simple. Of course you can add your own specific routes once created, but Aegis will not manage any of that.

That API Gateway resource then uses a Lambda proxy. This means the Lambda must return a specific JSON response that includes a status code, body, and headers.

On your side of things, you just need to run a few simple commands to setup a new Aegis project and then deploy. Aegis will run go build for you, zip your application up with a Node.js wrapper \(Go isn't natively supported by AWS Lambda, but that's not a reason to not use it\), configure the API Gateway, and upload the Lambda function.

Your Go application handles all of the incoming requests. It gets passed a JSON message and must return the proper JSON response. To make this easier, your application can use Aegis' lambda package to use a "router" of the sorts. This will parse the incoming messages and return the appropriate responses for you. As far as you're concerned, your code will look like it's using any other HTTP router for a RESTful API. Of course with some caveats. Lambda & API Gateway can't stream HTTP responses for starters.

That's the conventional part. However, Aegis is not devoid of configuration. You can work with a YAML file to adjust some things such as setting environment variables, setting up API stages, setting allocated Lambda resources, and more.

The rest of this guide will go over the finer details.

