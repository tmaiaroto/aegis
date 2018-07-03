# Secret

The `aegis secret` command is a _very important command._ It's a lightweight wrapper around AWS Secret Manager.
It will let you store and read values in AWS Secret Manager from your CLI. A nice feature about it is that
by default, `read`, will show parts of values hidden by asterisks.

The entire point is that you don't log or expose in the open (other people maybe looking at your screen),
sensitive credentials.

![Aegis secret read example](/aegis/img/aegis-cli-secret-read.png "Aegis secret read")

If you don't provide a key, it will display all values in a table.

Of course you can read the actual values as well if you use the `full` command instead of `read`.
It just makes you think first.

You can use `store` to store key values. You'll want to provide the secret name, key and then value.
There are some additional flags as well with this command. You can use custom KMS keys and more.

Like every other CLI command, there is help available from the CLI.

## How Aegis Uses Secrets

> aegis.yaml example (partial)

```
api:
  name: Example Cognito Aegis API
  description: An example API with auth using Cognito
  stages:
    prod:
      name: prod
      variables:
        poolid:
          key: PoolID
          value: <cognitoExample.PoolID>
        clientid:
          key: ClientID
          value: <cognitoExample.ClientID>
          # set the above using Aegis CLI, ex. 
```

Back to `deploy` here. Remember that when deploying, there's the `aegis.yaml` file. It is in here
that sensitive credentials (keys) in AWS Secret Manager are referenced.

You'll see in the Cognito example that <span class="nowrap">`<cognitoExample.PoolID>`</span> is used
for a <span class="nowrap">`poolid`</span> key that is used as an API Gateway stage variable. The value
for that key (and ultimately, stage variable) is pulled from AWS Secret Manager during deploy and used.

<aside class="note-info">
<i class="fas fa-info-circle"></i> At no point do sensitive credentials get logged. They should never exist
as plain text in configuration either. There is simply no way to accidentally publish them if using Aegis
as designed.
</aside>

An interesting thing about API Gateway stage variables...They do not support certain special characters.
AWS Lambda environment variables don't have as many character restrictions, but stage variables can be
problematic.

So what does Aegis do here? It base64 encodes the stage variables. You can decode them yourself in your
code or you can use Aegis' helpers to get the values.

Aegis has helpers and this concept of **_Aegis Variables_** which is really just a fancy way of saying,
go check:

 * AWS Lambda environment variables
 * API Gateway stage variable
 * actual environment variables

The helper tries a few places to find the variable you're after and it will then ensure the string
is base64 decoded if it was encoded at all. It works regardless.

[See the Aegis Variables section](/aegis/helpers/#aegis-variables) for more information about the helper function and where these
variables are available within your application.