# Log

```go
logger := logrus.New()
// Various configuration options for logrus...

AegisApp = aegis.New(handlers)
// Set your configured logger on Log
AegisApp.Log = logger
AegisApp.Start()
```

Aegis' Log dependency provides an adaptable way to log using the popular Go package
<a href="https://github.com/sirupsen/logrus" target="_blank">logrus</a>.
You're certainly welcome to use the standard library `log` package as well. AWS Lambda will send all
of that stdout to CloudWatch for you automatically.

However, there may be other logging services you wish to use. Rather than re-invent the wheel, Aegis
used logrus instead because of its adaptability.

Set the `Log` field on the `Aegis` interface like any other core dependency to change it. You can
also use the <span class="nowrap">`ConfigureLogger(*logrus.Logger)`</span> helper function.

Do note that everything internally will use `log`. So anything that Aegis logs out will be found in
CloudWatch. It's not that Aegis didn't want to dogfood its own logging dependency, it's that Aegis
did not want its logging to get in the way of your logging.

For example, if your handler (just one of your Lambdas) wanted to use `Log` to send information
to your team Slack channel, you wouldn't also want internal Aegis logging to be sent as well. Especially
because you couldn't turn them off. There is currently no sort of verbosity setting in Aegis. If you
were to, say, disable CloudWatch logging for your Lambda altogether then the standard `log` calls
would simply go no where.

CloudWatch of course isn't the prettiest way to view your logs. While logrus has fancy coloring,
CloudWatch is black and white text through your web browser in the AWS Console. So you might end up
configuring `Log` in a different way to work with logging in a prettier way.

Another thing worth noting here is that logrus' 3rd party formatters can be rather powerful. There are
formatters for things like fluentd, logstash, and more. So you might just find yourself using it for
more than simply basic logging. Also keep in mind that AWS has a hosted version of Elasticsearch with
Kibana...See where this is going?? You can implement a rather robust searchable logging solution all
within AWS.