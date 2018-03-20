package framework

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// Tasker struct provides an interface to handle scheduled tasks
type Tasker struct {
	handlers            map[string]TaskHandler
	IgnoreFunctionScope bool
}

// TaskHandler is similar to RouteHandler except there is no response or middleware
type TaskHandler func(context.Context, *map[string]interface{}) error

// LambdaHandler is a native AWS Lambda Go handler function. Handles a CloudWatch event.
func (t *Tasker) LambdaHandler(ctx context.Context, evt map[string]interface{}) error {
	var err error

	handled := false
	taskName := ""
	if name, ok := evt["_taskName"]; ok {
		taskName = name.(string)
	}

	// If there's a _taskName, use the registered handler if it exists.
	if handler, ok := t.handlers[taskName]; ok {
		handled = true
		// Capture the handler in XRay automatically
		err = xray.Capture(ctx, "TaskHandler", func(ctx1 context.Context) error {
			// Annotations can be searched in XRay.
			// For example: annotation.TaskName = "mytask"
			xray.AddAnnotation(ctx1, "TaskName", taskName)
			xray.AddMetadata(ctx1, "TaskEvent", evt)
			return handler(ctx, &evt)
		})
	}
	// Otherwise, use the catch all (router "fallthrough" equivalent) handler.
	// The application can inspect the map and make a decision on what to do, if anything.
	// This is optional.
	if !handled {
		// It's possible that the Tasker wasn't created with NewTasker, so check for this still.
		if handler, ok := t.handlers["*"]; ok {
			// Capture the handler in XRay automatically
			err = xray.Capture(ctx, "TaskHandler", func(ctx1 context.Context) error {
				xray.AddAnnotation(ctx1, "TaskName", taskName)
				xray.AddAnnotation(ctx1, "FallthroughHandler", true)
				xray.AddMetadata(ctx1, "TaskEvent", evt)
				return handler(ctx, &evt)
			})
		}
	}

	return err
}

// Listen will start a task listener which acts much like a router except that it handles scheduled task events instead
func (t *Tasker) Listen() {
	lambda.Start(t.LambdaHandler)
}

// NewTasker simply returns a new Tasker struct and behaves a bit like Router, it even takes an optional rootHandler or "fall through" catch all
func NewTasker(rootHandler ...TaskHandler) *Tasker {
	// The catch all is optional, if not provided, an empty handler is still called, but nothing happens.
	handler := func(context.Context, *map[string]interface{}) error { return nil }
	if len(rootHandler) > 0 {
		handler = rootHandler[0]
	}
	return &Tasker{
		handlers: map[string]TaskHandler{
			"*": handler,
		},
	}
}

// Handle will register a handler for a given task name
func (t *Tasker) Handle(name string, handler TaskHandler) {
	if t.handlers == nil {
		t.handlers = make(map[string]TaskHandler)
	}
	t.handlers[name] = handler
}
