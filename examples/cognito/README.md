## Aegis Cognito Protected API Example
----------------------

This example shows how to set up authentication using Cognito for an Aegis create API.
Aside from having an AWS account, the only pre-requisite here is that you also have
a Cognito User Pool set up with an app client and domain.

For actual front-end JavaScript apps, take a look at AWS Amplify: 
https://github.com/aws/aws-amplify

### Cognito User Pool Requirements

We're just using a very basic user pool. The user attributes don't matter.
Though you'll want to set up a domain name so you can redirect to your AWS Cognito hosted
login/signup page.

You'll also need a app client created. The callback URL(s) should include your 
API Gateway endpoint with callback path. So something like:

`https://123jkldsf.execute-api.us-east-1.amazonaws.com/prod/callback`

The sign out URL doesn't matter in this case.

This means you will need to deploy the Aegis app in a non-working state at first.
It won't work because the Cognito app client will not have your valid callback URL.
It's just a chicken and egg scenario.

Once you deploy, take the API Gateway domain and put `/callback` on the end
and put it into your Cognito app client's configuration in the AWS web console.

Note that the code automatically handles the API domain name, so you won't need
to change anything there for redirects, etc. In real production code you might
configure that somewhere to use and remove all the code here in this example that
looks at the `evt` and headers and such.

Ensure your app client's OAuth 2.0 section has "Authorization code grant" checked.
The OAuth scopes don't matter.

### Edits to the example code

You only need to edit three strings to make the example work and it's commented
in the code. It's the configuration for the Cognito app client. Maybe you'll only
need to edit two of those if you're deploying in the us-east-1 region.

```
Region:   "us-east-1",
PoolID:   "us-east-1_Xxxxx",
ClientID: "xxxxx",
```

Alternatively, you could use Lambda environment variables set in `aegis.yaml`
or API Gateway stage variables also set there. You likely wouldn't want to hard
code these values into your production application. Though note that the Cognito
app client helper does not need the secret key. Since your Lambda is authorized
to make Cognito SDK calls, it will retrieve it automatically. This reduces your
security risks with accidentally published credentials.

### Deploying

Just run `aegis deploy` in this directory and you should be good to go.
Note that you can edit the `aegis.yaml` file to rename the Lambda function, API,
or make any other adjustments as desired.