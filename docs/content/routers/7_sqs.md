# SQS Router

```go
func main() {
    sqsRouter := aegis.NewSQSRouterForBucket("aegis-queue")
    sqsRouter.Handle("attr", "value", handleSQSMessage)

    handlers := aegis.Handlers{
        SQSRouter: sqsRouter,
    }
}

func handleSQSMessage(ctx context.Context, d *aegis.HandlerDependencies, evt *aegis.SQSEvent) error {
    // evt will contain the message body, attributes, etc.
    return nil
}
```

Each SQS queue allows one Lambda to be invoked when new messages enter the queue. However, multiple queues may go
to trigger the same Lambda function. This router plays an important role here.

Further, SQS messages can vary wildly. You will end up defining your message conventions. This is especially true
when it comes to each message's attributes. This is where the router's rules are applied too.

When you handle an SQS event you can match on a message attribute name and it's string value. SQS will always
provide a string representation of the message attribute value. If it is binary, this means a base64 string.
Therefore `Handle()` always takes two strings for the first two arguments; key and value.

If you wish to handle all messages for a given queue, where you have the most flexibility, you can simply use
an SQS router's root/falltrhough handler. Like the S3Object Router, there is a more terse <span class="nowrap">`NewSQSRouterForQueue()`</span>
function to create a router for a specific queue in one line.

**Note: At this time Aegis CLI will not create or manage SQS queues for you.** It will associate the necessary
execution role for your Lambda, but you will need to manage your own queues and which Lambda functions they trigger.