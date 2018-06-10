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

