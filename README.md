# Aegis

[![License Apache 2](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://github.com/tmaiaroto/aegis/blob/master/LICENSE) [![godoc aegis](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/tmaiaroto/aegis) [![Build Status](https://travis-ci.org/tmaiaroto/aegis.svg?branch=master)](https://travis-ci.org/tmaiaroto/aegis) [![Go Report Card](https://goreportcard.com/badge/github.com/tmaiaroto/aegis)](https://goreportcard.com/report/github.com/tmaiaroto/aegis)

**[Aegis Documentation](https://tmaiaroto.github.io/aegis/)**

Aegis is both a simple deploy tool and framework. Its primary goal is to help you write
microservices in the AWS cloud quickly and easily. They are mutually exclusive tools.

Aegis is not intended to be an infrastructure management tool. It will never be
as feature rich as tools like [Terraform](https://www.terraform.io). Its goal is
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

#### Building

It's easiest to download a binary to use Aegis, though you may wish to build for your specific platform. 
In this case, Go Modules is used. Easiest thing to do after cloning is:

```GO111MODULE=on go mod download```

Then build:

```GO111MODULE=on go build```

Unfortunately you can't do a straight `go build` because of one of the packages used. You'll get errors.
So using Go Modules is the way.

#### Contributing

Please feel free to contribute (see CONTRIBUTING.md). Though outside of actual pull requests with code,
please file issues. If you notice something broken, speak up. If you have an idea for a feature, put it
in an issue. Feedback is perhaps one of the best ways to contribute. So don't feel compelled to code.

Keep in mind that not all ideas can be implemented. There is a design direction for this project and
only so much time. Though it's still good to share ideas.

#### Running Tests

Goconvey is used for testing, just be sure to exclude the `docs` directory. For example: `goconvey -excludedDirs docs`

Otherwise, tests will run and also include the `docs` folder which will likely have problems.

Alternatively, run tests from the `framework` directory.
