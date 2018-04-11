# Aegis Framework Examples
---------------------

The best way to test and document is through extensive usage and example. These examples aim to provide
help beyond documentation and in some cases before being written into documentation.

### Examples Relying Upon Sensitive Credentials

Some of these examples use services which require sensitive credentials.
They have been omitted from the repository for obvious reasons.
You would need to fill in the blanks with your own information to run those examples.

How do you handle sensitive credentials? A few strategies (I'm sure you'll think of more):

 - Put them in aegis.yaml, it sets them on API Gateway stage variables or Lambda environment variables.
   Just be sure to not commit that aegis.yaml to any public repository. Safe guard it.

 - Pull them from a file in S3 upon startup. This fetch will not occur on a "warm" Lambda invocation.
   Remember, your Lambda has access to read from your S3 without providing any credentials in your code.

 - Use a service like etcd that is behind a firewall. It'd be similar to the S3 approach, only more 
   flexible and perhaps faster.

 - Manually edit your Lambda environment variables or API Gateway stage variables after you
   deploy your Lambda.
   
 - Use the framework, but not the deploy tool. Get creative with ldflags and set variables
   from your CLI upon running `go build`. Zip your own binary and upload to Lambda.

Have an idea that fits into Aegis? Share it in the GitHub issues as an enhancement.
None of the above solutions are perfect.