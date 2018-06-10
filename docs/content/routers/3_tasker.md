# Tasker

The `Tasker` is an interface to route scheduled jobs or "tasks" that your Lambda may perform. This works
in conjunction with CloudWatch Rules. So you'll need to have events setn to your Lambda from CloudWatch.

```json
{
    "schedule": "rate(1 minute)",
    "disabled": true,
    "input": {
        "_taskName": "test",
        "foo": "bar"
    }
}
```

Aegis makes this easy to do through a conventional approach. You need not configure CloudWatch events
manually (though you can of course). You also don't add them to your `aegis.yaml` or anything like that.
It's much easier; you'll just add a `tasks` directory at the root of your source code where you run
`aegis deploy`. Within this directory, you can include JSON files for each scheduled task you'd like
to set up.

### Task JSON Definition Format

The first key being `schedule` which takes a CloudWatch Event Rule expression (ie. "rate" or crontab format).
The optional `disabled` key in your JSON task definition can disable your CloudWatch rule so you don't
need to delete it just to turn it off for a while.

Then you have the `input` key which contains an object that has your event message. This can be anything
you want. It comes in to your handler as a `map[string]interface{}` and is "static JSON" in terms of
the CloudWatch Rule configuration.

There is a conventional `_taskName` key in your payload that you need to be aware of. This is what gets
matched by the `Tasker` in order to delegate to your handlers.

<aside class="note-info">
<i class="fas fa-info-circle"></i> Note that tasks are routed by exact match.
</aside>

```go
tasker := aegis.NewTasker(taskerFallThrough)
tasker.Handle("test", handleTestTask)
```

> The handler should look familiar, excpet that the event is simply a map

```go
// Example task handler
func handleTestTask(ctx context.Context, d *aegis.HandlerDependencies, evt map[string]interface{}) error {
	log.Println("Handling task!", evt)
	return nil
}
```

### Handling Tasks

Like the API Gateway Proxy Request `Router`, `Tasker` has a "fall through" or "catch all" as well.
It's function signature is no different than any other task handler. This fall through handler is also
optional. You can call `aegis.NewTasker()` to use just as well. In that case, if your configured `Tasker`
does not match any task names, it will not route those events and they won't be handled.

Obviously, there is no one to return a response message to in the case of a scheduled job. An `error`, or `nil`,
must still be returned. Though the error would only be for the benefit of your tracing tool (ie. X-Ray will see it).

The event in this case will be whatever you defined in the definition JSON. Often times you may find that the
most important thing is the specific handler function itself is called and not so much this static JSON payload.