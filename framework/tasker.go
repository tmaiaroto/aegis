package framework

import (
	"bytes"
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

// Tasker struct provides an interface to handle scheduled tasks
type Tasker struct {
	handlers            map[string]TaskHandler
	IgnoreFunctionScope bool
}

// TaskHandler is similar to RouteHandler except there is no response or middleware
type TaskHandler func(context.Context, *CloudWatchEvent)

// LambdaHandler is a native AWS Lambda Go handler function. Handles a CloudWatch event.
func (t *Tasker) LambdaHandler(ctx context.Context, evt CloudWatchEvent) {
	// Only handle scheduled events
	if evt.Source == "aws.events" && evt.DetailType == "Scheduled Event" {
		// Figure out which handler to use.
		if len(evt.Resources) > 0 {
			// Would there ever be more than one??
			// "resources": [ "arn:aws:events:us-east-1:123456789012:rule/MyScheduledRule" ],
			// Aegis events are always JSON and are defined in JSON files along with configuration.
			// Their ARN/IDs ultimately become something like:
			// arn:aws:events:us-east-1:1234567890:rule/aegis_aegis.example.json
			name := ""
			parts := strings.Split(evt.Resources[0], "/")
			if len(parts) > 1 {
				// Glue together <function name>.<event file name>
				// So not only does the handler handle by file name, but also function name.
				// This means "event.json" from one Aegis created Lambda won't conflict with
				// another "event.json" created by another Aegis Lambda.
				// UNLESS... IgnoreFunctionScope was set to true.
				// Then any simple name match will work. Good for local testing. Good for other cases.
				// But tasker.Handle(name, ...) need not define the function name. It can just use the json file name.
				var buffer bytes.Buffer
				if !t.IgnoreFunctionScope {
					// Protect in case the tasker.Handle(name, ...) did include the function name.
					if !strings.Contains(parts[1], lambdacontext.FunctionName) {
						buffer.WriteString(lambdacontext.FunctionName)
						buffer.WriteString(".")
					}
				}
				buffer.WriteString(parts[1])
				name = buffer.String()
				buffer.Reset()
			}
			if handler, ok := t.handlers[name]; ok {
				handler(ctx, &evt)
			}
		}
	}
}

// Listen will start a task listener which acts much like a router except that it handles scheduled task events instead
func (t *Tasker) Listen() {
	lambda.Start(t.LambdaHandler)
}

// NewTasker simply returns a new Tasker struct and behaves a bit like Router
func NewTasker() *Tasker {
	return &Tasker{}
}

// Handle will register a handler for a given task name
func (t *Tasker) Handle(name string, handler TaskHandler) {
	if t.handlers == nil {
		t.handlers = make(map[string]TaskHandler)
	}
	t.handlers[name] = handler
}
