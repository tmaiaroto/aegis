# Deploy

The `aegis deploy` command will likely be the next command you're after. Running it anywhere there is buildable
Go code (that handle Lambda events of course) as well as an `aegis.yaml` will result in a deployment to AWS.

This includes building your Go binary, uploading it to AWS, and creating resources in AWS.

Again, Aegis' CLI tools and the `framework` package are mutually exclusive. Deploying a Lambda that handles
events using the vanilla `aws-lambda-go` package from Amazon should work just fine. The CLI tools and framework
are designed to work together of course, so you'll get the most benefit from using both.

Upon deploying, Aegis will create an Aegis IAM role (if it doesn't already exist and if you didn't specify to
use a different IAM role in your `aegis.yaml` config). It's someone permissive by default because there's no
introspection into your code to see what you're doing.

In the future, there is some intent for auditing and reporting back a summary of permissions. However,
this command is mainly intended to be a development tool -- not part of a production work flow. You might
reach for other tools that let your IAM roles and other cloud resources remain under revision control.
For example Terraform is a great tool in this case.

However, Aegis does consider DevSecOps, so there are some considerations around security. For example,
when the deploy command runs, it will look into AWS Secret Manager in order to get sensitive values to
configure in Lambda environemt variables or API Gateway stage variables.

The deploy command will also report back some helpful information to your console as it goes along
creating resources.