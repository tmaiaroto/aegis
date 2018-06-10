# S3 Object Router

```go
func main() {
    s3Router := aegis.NewS3ObjectRouterForBucket("aegis-incoming")
    // s3Router.Handle("s3:ObjectCreated:Put", "*.png", handleS3Upload)
    // Put() is a shortcut for the above
    s3Router.Put("*.png", handleS3Upload)

    handlers := aegis.Handlers{
        S3ObjectRouter: s3Router,
    }
}

func handleS3Upload(ctx context.Context, d *aegis.HandlerDependencies, evt *aegis.S3Event) error {
    // evt will contain bucket name, path, etc.
    return nil
}
```

Now here we have an interesting Router. Like the API Gateway Proxy, we have a few things to match on.
It's not just a path, but also a method or operation. In the case of S3 objects, we're talking about
puts and deletes, other such operations, and now also a bucket (or domain if you like).

So the `S3ObjectRouter` has some convenient methods on it for you to use; `Put()`, `Delete()`, `Copy()`,
and so on. Though you can choose to write it the "long" way and use `Handle()` where the first argument
would be the actual S3 object event name.

There are two "new" methods for this router. <span class="nowrap">`NewS3ObjectRouter()`</span> and
<span class="nowrap">`NewS3ObjectRouterForBucket()`.</span> Both will take an optional "fall through"
handler. Both use the same router interface. <span class="nowrap">`NewS3ObjectRouterForBucket("bucket-name")`</span>
takes a bucket name and sets it on to the router interface. <span class="nowrap">`NewS3ObjectRouter()`</span>
does not.

This means that you can not only "catch all" operations and objects, but also all operations and object
for any bucket that's emitting events that trigger your Lambda.

Of course, you could always switch the bucket the router is working with after your have the interface.
For example, assumimg you have <span class="nowrap">`r := NewS3ObjectRouter()`</span>, you could then do
<span class="nowrap">`r.Bucket = "my-bucket"`.</span>

## Configuring S3 Object Events

```yaml
bucketTriggers:
  - bucket: aegis-incoming
    filters:
      - name: suffix
        value: png
      #- name: prefix
      #  value: path/
    eventNames:
      - s3:ObjectCreated:*
      - s3:ObjectRemoved:*
      # ... there's a few and there's wildcards, see:
      # https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations
    disabled: false
```

Unlike tasks or RPCs, we can't simply conventionally catch events that might come our way. We actually need to
ensure these events are being sent. The way that happens is through S3 object event triggers. You can set these
up yourself of course (or using your favorite tool like Terraform or CloudFormation). Or, you could leverage
`aegis deploy` by configuring your `aegis.yaml` file.

Aegis will look for this configuration each time you call `aegis deploy` from the CLI. You can also disable the
triggers using the `disabled` key. Set that to true and deploy again. Re-enable when you're ready again. This way
you don't need to lose anything in your configuration. Aegis is also non-destructive.

<aside class="note-warning">
<i class="fas fa-exclamation-triangle"></i> Note that Aegis will not know to delete your S3 object event triggers.
It can disable them, but it will not remove them. So you will need to manually do so yourself if you no longer want them.
</aside>

## Event Matching

> What goes on within Aegis to match an S3 object event

```go
if r.Bucket == "" || r.Bucket == record.S3.Bucket.Name {
    // Handlers are registered in a map.
    // The keys in this map contain your match pattern string.
    for globStr, handler := range r.handlers {
    g = glob.MustCompile(globStr)
    if g.Match(record.S3.Object.Key) {
        // ...Goes on to handle the event, calling your handler.
    }
}
```

Unlike most other routers, S3 object events are matched by glob match. Again, if no match is found, it will use a
fall through handler if you provided one. Note that the bucket name is optional. You could handle objects sent to
any bucket provided its events trigger your Lambda.

So for example, `*.png` will handle any object that ends with a png extension. Whereas `{*.png,*.gif,*.jpg}` would handle
any file with those extensions. The package used for glob matching is [gobwas/glob](https://github.com/gobwas/glob).
You can read its documentation for all the options you have available to match, it's rather robust.

A good thing to think about here is the order in which you define your routing rules. It won't really be a concern
for you with other routers, but this one can result in multiple rules matching the same event depending on what
you have set up. The first route match found will be used, calling its handler.

Remember, there are 3 things at play here:

 * The bucket name (if not set, any bucket is a match)
 * The S3 object event name (if not set, any event is a match)
 * The S3 object key name (glob match)