// Copyright © 2016 Tom Maiaroto <tom@SerifAndSemaphore.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package framework

import (
	"context"
	"errors"

	"github.com/aws/aws-lambda-go/lambda"
)

// Tasker struct provides an interface to handle scheduled tasks
type Tasker struct {
	handlers            map[string]TaskHandler
	IgnoreFunctionScope bool
	Tracer              TraceStrategy
}

// TaskHandler is similar to RouteHandler except there is no response or middleware
type TaskHandler func(context.Context, *HandlerDependencies, map[string]interface{}) error

// LambdaHandler is a native AWS Lambda Go handler function. Handles a CloudWatch event.
func (t *Tasker) LambdaHandler(ctx context.Context, d *HandlerDependencies, evt map[string]interface{}) error {
	// If an incoming event can be matched to this router, but the router has no registered handlers
	// or if one hasn't been added to aegis.Handlers{}.
	if t == nil {
		return errors.New("no handlers registered for Tasker")
	}
	var err error

	handled := false
	taskName := ""
	if name, ok := evt["_taskName"]; ok {
		taskName = name.(string)
	}

	if t.handlers != nil {
		// If there's a _taskName, use the registered handler if it exists.
		if handler, ok := t.handlers[taskName]; ok {
			handled = true
			// Trace (defeault is to use XRay)
			t.Tracer.Record("annotation",
				map[string]interface{}{
					"TaskName": taskName,
				},
			)
			t.Tracer.Record("metadata",
				map[string]interface{}{
					"TaskEvent": evt,
				},
			)

			err = t.Tracer.Capture(ctx, "TaskHandler", func(ctx1 context.Context) error {
				d.Tracer = &t.Tracer
				return handler(ctx1, d, evt)
			})
		}
		// Otherwise, use the catch all (router "fallthrough" equivalent) handler.
		// The application can inspect the map and make a decision on what to do, if anything.
		// This is optional.
		if !handled {
			// It's possible that the Tasker wasn't created with NewTasker, so check for this still.
			if handler, ok := t.handlers["_"]; ok {
				// Trace (defeault is to use XRay)
				t.Tracer.Record("annotation",
					map[string]interface{}{
						"TaskName":           taskName,
						"FallthroughHandler": true,
					},
				)
				t.Tracer.Record("metadata",
					map[string]interface{}{
						"TaskEvent": evt,
					},
				)

				err = t.Tracer.Capture(ctx, "TaskHandler", func(ctx1 context.Context) error {
					d.Tracer = &t.Tracer
					return handler(ctx, d, evt)
				})

			}
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
	handler := func(context.Context, *HandlerDependencies, map[string]interface{}) error { return nil }
	if len(rootHandler) > 0 {
		handler = rootHandler[0]
	}
	return &Tasker{
		handlers: map[string]TaskHandler{
			"_": handler,
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
