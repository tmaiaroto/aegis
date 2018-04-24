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

### Credentials/code edits

There are only three strings you need to immediately be concerned with in order
to make this example work. The code needs a `region`, `pool ID`, and `client ID`
from Cognito. Aegis' CLI has a `secret` command that can help you manage these
sensitive credentials (and other key/value settings) so you don't need to hard
code anything (which we would never do, right? right??). The code is set up to
read these variables referred to as "Aegis variables." Really, they are just
either Lambda environment variables or API Gateway stage variables. You can
use either.

In order to set these values, you'll want to use the `secret` command and
set corresponding values in the `aegis.yaml`. Open the `aegis.yaml` and look
at the API Gateway stage variables section. You'll see values like:
`<cognitoExample.PoolID>`

These denote the use of AWS Secrets manager and upon Aegis `deploy`, they
will be retrieved and set as Lambda environment variables or API Gateway
Stage variables as per the config.

If you'd like to just use what's in the config already, run the following
commands and provide your own Cognito Pool values.

`aegis secret store cognitoExample PoolID xxxxxx`

`aegis secret store cognitoExample ClientID xxxxxx`

The region was actually hard coded in this example because it's not a sensitive
piece of information, though you can change it if you like. There's a helper
function that reads the variables no matter if they are Lambda environment
variables or API Gateway stage variables (which take priority).

Note that in order to update these values, you must re-deploy. However,
you could also look at AWS Secrets Manager yourself manually in your code
and use fresh values directly from it of course. Just keep in mind that's
an extra request that has to complete before sending a response back to
the client. It's much faster to read environment/stage variables.

Alternatively, if you want to form a bad habit (or save 40 cents for the month?), 
you could simply edit three lines in the example code to make it work. Maybe 
you'll only need to edit two of those if you're deploying in the us-east-1 region.

```
Region:   "us-east-1",
PoolID:   "us-east-1_Xxxxx",
ClientID: "xxxxx",
```

Though note that the Cognito app client helper does not need the secret key. 
Since your Lambda is authorized to make Cognito SDK calls, it will retrieve 
it automatically. This reduces your security risks with accidentally published 
credentials to a degree. However, it is reccommended to use the AWS Secrets
Manager along with setting the variables in `aegis.yaml` config to be stored
on Lambda environment variables or API Gateway stage variables -- both of
which are encrypted when transmitted over the wire.

### Deploying

Just run `aegis deploy` in this directory and you should be good to go.
Note that you can edit the `aegis.yaml` file to rename the Lambda function, API,
or make any other adjustments as desired.