# SES Router

```go
func main() {
    sesReceiver := aegis.NewSESRouter()
    sesReceiver.Handle("*@ses.serifandsemaphore.io", handleEmail)

    handlers := aegis.Handlers{
        SESRouter: sesReceiver,
    }
    AegisApp = aegis.New(handlers)
    AegisApp.Start()
}

func handleEmail(ctx context.Context, d *aegis.HandlerDependencies, evt *aegis.SimpleEmailEvent) error {
    log.Println(evt)
    return nil
}
```

The AWS SES (Simple E-mail Service) router, `SESRouter`, and the ability for a Lambda to receive e-mail
is perhaps one of the coolest things ever. Yes, there is such a thing as serverless e-mail.

You can route based on any address match, or address matches within a specific domain, using
<span class="nowrap">`NewSESRouter()`</span> and <span class="nowrap">`NewSESRouterForDomain()`</span> respectively.

Like the `S3ObjectRouter`, the SES router will use glob based matching. So, first the domain is check if set,
and then your pattern is checked against the e-mail address.

## How to Read E-mails

> Example SES config

```yaml
sesRules:
  - ruleName: aegis-test
    enabled: true
    requireTLS: false
    scanEnabled: true
    # ruleSet: ... optional string, defaults to aegis-default-rule-set
    # invocationType: Event # default to Event, RequestResponse is other option less common and has to return in 30 seconds
    # snsTopicArn: ... optional SNS topic to also notify - SNS setup is outside aegis deploy for now
    recipients:
      - example@ses.serifandsemaphore.io
    # Also, an S3 bucketTrigger could be used to pick up the e-mail in its entirety OR the an SES event can be
    # used because it will provide a message ID that can be looked for in S3.
    s3Bucket: "aegis-incoming"
    # subdirectory
    s3ObjectKeyPrefix: "ses_"
    s3encryptMessage: false
    # s3KMSKeyArn: ... s3encryptMessage will use the default KMS key for encrypting email if true unless this is provided
    # s3SNSTopicArn: "sometopic" ... an optional SNS topic to publish when messages are saved to S3, different from the SNS topic about the event
```

There's something important to note about what you get in the event message from an SES event. You are not
getting the full body of the e-mail. Just the e-mail "headers" come through. Things like the from address,
to address, CC, BCC, and subject line.

If you want to actually read the body of the e-mail, you'll need to set up a few things in AWS first. What
happens is the e-mail gets stored, in full, into S3 (if it's set up to do so). You then read it from S3.
E-mails can be stored in S3 encrypted too if you configure that as well.

To help with all this configuration in AWS, Aegis `deploy` command will read settings from `aegis.yaml` to
set things up.

Also keep in mind that you can use an `S3ObjectRouter` to handle incoming objects into whatever bucket is
receiving your e-mails too. The difference is that you wouldn't then get the e-mail headers and you'd _have_
to read the file to know what the e-mail was about. The SES handler would let your code make a determination
as to whether or not it wanted to retrieve the file from S3.



