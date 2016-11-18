Aegis
==========

A simple utility for deploying a Golang based Lambda with an API using 
AWS API Gateway's `ANY` method with a `{proxy+}` path to handle any request.

This results in a very easy solution for building serverless APIs with Go.

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

### Project Name

Curious why the name Aegis? The "shield" metaphore comes from putting up an application that's 
protected from downtime as Lambda scales very well. It's also Athena's shield who is the goddess
of crafts and wisdom. So now go craft something clever and enjoy!