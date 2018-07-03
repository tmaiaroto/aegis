# Update

The `update` command is similar deploy, but it just updates the Lambda function itself.

A full `deploy` will set create resources and configure things in AWS. While an `update`
will only build your Go binary, zip, and then upload to AWS Lambda.

It saves a little bit of time when you have some code changes because it doesn't
need to check for existing resources using the AWS SDK which can make a few HTTP
requests.

More than likely the slowest part about deployment will be the time it takes to upload
the binary. However, this command is still available if you prefer.

