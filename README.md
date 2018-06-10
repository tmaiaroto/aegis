# Aegis

[![License Apache 2](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://github.com/tmaiaroto/aegis/blob/master/LICENSE) [![godoc aegis](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/tmaiaroto/aegis) [![Build Status](https://travis-ci.org/tmaiaroto/aegis.svg?branch=master)](https://travis-ci.org/tmaiaroto/aegis) [![Go Report Card](https://goreportcard.com/badge/github.com/tmaiaroto/aegis)](https://goreportcard.com/report/github.com/tmaiaroto/aegis)

Aegis is both a simple deploy tool and framework. It's primary goal is to help you write
microservices in the AWS cloud quickly and easily. They are mutually exclusive tools.

Aegis is not intended to be an infrastructure management tool. It will never be
as feature rich as tools like [Terraform](https://www.terraform.io). It's goal is
to assist in the development of microservices - not the maintenance of infrastructure.

Likewise the framework is rather lightweight as well. It may never have helpers and
features for every AWS product under the sun. It provides a conventional framework
to help you build serverless microservices faster. It removes a lot of boilerplate.

### Getting Started

You'll need an AWS account of course. You'll also want to have your credentials setup
as you would for using AWS CLI. Note that you can also pass AWS credentials via the 
CLI or by setting environment variables.

Get Aegis of course. Use the normal `go get github.com/tmaiaroto/aegis`.
Ensure the `aegis` binary is in your executable path. You can build a fresh copy
from the code in this repository or download the binary from the releases section
of the GitHub project site. If you want to use the framework though, you'll need to
use go get anyway.

You can find some examples in the `examples` directory of this repo. Aegis also comes
with a command to setup some boilerplate code in a clean directory using `aegis init`.
Note that it will not overwrite any existing files.

Work with your code and check settings in `aegis.yaml`. When you're ready, you can deploy
with `aegis deploy` to upload your Lambda and setup some resources.

Aegis' deploy command will set up the Lambda function, an optional API Gateway, IAM roles,
CloudWatch event rules, and other various triggers and permissions for your Lambda function.
You're able to choose a specific IAM role if you like too. Just set it in `aegis.yaml`.

If you're deploying an API, the CLI output will show you the URL for it along with other
helpful information.

The Aegis framework works by handling events (how anything using AWS Lambda works). The way
in which it does this though is via "routers." This means your Lambda is actually able to
handle multiple types of events if you so choose.

Many people will want to write one handler for one Lambda, but that's not a mandate of Lambda.
So feel free to architect your microservices how you like.

There are several types of routers. You can handle incoming HTTP requests via API Gateway using
various HTTP methods and paths. You can handle incoming S3 events. You can handle scheduled Lambda
invocations using CloudWatch rules. You can even handle invocations from other Lambdas ("RPCs").

#### Logging

Go's normal logging will work and end up in CloudWatch. Additionally, logrus is available under the `Aegis.Log`.
You'll need to set up a new Aegis interface with `framework.New()`. For example:

```
import (
    aegis "github.com/tmaiaroto/aegis"
)

func main() {
    app := aegis.New(&aegis.Handlers{})
    app.Log.Println("log stuff")
    app.Log.Error(errors.New("bad stuff"))
}
```

Also note that `Log` is injected into each handler's dependencies. So you can pick it up from the second
argument, right after context.

Logrus was chosen here. So you can configure the interface with any instance of Logrus and you can use any
plugin for Logrus. Send logs to your Slack for fun. Go nuts.

```
func handler(ctx context.Context, d *aegis.HandlerDependencies, evt map[string]interface{}) {
    d.Log.Println("go nuts")
}
```

All internal framework logs use standard Go `log` and will end up in CloudWatch.

#### Tracing

Aegis uses AWS XRay by default, though you can change the tracing strategy. You just need to implement
the interface. All event handlers are automatically traced with annotations and metadata if applicable.

You are able to add to these annoations and metadata. You are also free to use XRay in your handlers
yourself. A `Tracer` will be injected into each handler.

#### Contributing

Please feel free to contribute (see CONTRIBUTING.md). Though outside of actual pull requests with code,
please file issues. If you notice something broken, speak up. If you have an idea for a feature, put it
in an issue. Feedback is perhaps one of the best ways to contribute. So don't feel compelled to code.

Keep in mind that not all ideas can be implemented. There is a design direction for this project and
only so much time. Though it's still good to share ideas.

#### Running Tests

Goconvey is used for testing, just be sure to exclude the `docs` directory. For example: `goconvey -excludedDirs docs`

Otherwise, tests will run and also include the `docs` folder which will likely have problems.