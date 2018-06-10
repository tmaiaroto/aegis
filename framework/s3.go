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
	"github.com/gobwas/glob"
)

// S3ObjectRouter struct provides an interface to handle S3 object events (routers can be for a specific bucket or all buckets)
// https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example-configure-event-source.html
type S3ObjectRouter struct {
	handlers map[string]S3ObjectHandler
	Bucket   string
	Tracer   TraceStrategy
}

// S3ObjectHandler handles routed events in a different way than other handlers
// The router used for HTTP request like events is similar because it deals with http methods
// and paths, but not file extensions...Also, S3 events are slightly different than http methods.
// So this handler actually is a struct with a Handler function like normal, but then also
// an event type match and an object key ("file path") glob matcher.
type S3ObjectHandler struct {
	Handler func(context.Context, *HandlerDependencies, *S3Event) error
	Event   string
	// Key     string <-- not needed, the router's handlers map has the key match in its keys
}

// LambdaHandler handles S3 events.
func (r *S3ObjectRouter) LambdaHandler(ctx context.Context, d *HandlerDependencies, evt S3Event) error {
	// If an incoming event can be matched to this router, but the router has no registered handlers
	// or if one hasn't been added to aegis.Handlers{}.
	if r == nil {
		return errors.New("no handlers registered for S3ObjectRouter")
	}
	var err error
	handled := false

	// There can apparently be multiple, a router handler is to handle one object/file operation
	for _, record := range evt.Records {
		// log.Println("Looking to handle first record:", record)

		// If there are any handlers registered
		if r.handlers != nil {
			// and if the bucket matches or if no bucket was defined for the S3ObjectRouter (all buckets)
			if r.Bucket == "" || r.Bucket == record.S3.Bucket.Name {

				// Each key on handlers map is a string that can be used to glob match
				// If one of them matches, use that
				for globStr, handler := range r.handlers {
					// and also if the event matches or if no event was defined for the S3ObjectHandler (all event sources)
					// and of course this also matches by glob match too with values like `s3:ObjectCreated:*` and `s3:ObjectCreated:Put`
					esG, err := glob.Compile(handler.Event)
					if handler.Event == "" || (err == nil && esG.Match(record.EventSource)) {
						g, err := glob.Compile(globStr)
						if err == nil && g.Match(record.S3.Object.Key) {
							handled = true
							// NOTE: If Tracer is nil, we have a problem.
							r.Tracer.Record("annotation",
								map[string]interface{}{
									"S3Bucket":    record.S3.Bucket.Name,
									"S3ObjectKey": record.S3.Object.Key,
									"S3Event":     record.EventName,
								},
							)
							err = r.Tracer.Capture(ctx, "S3ObjectHandler", func(ctx1 context.Context) error {
								d.Tracer = &r.Tracer
								return handler.Handler(ctx1, d, &evt)
							})
						}
					}
					// TODO: think about some verbose setting for the framework.
					// Possibly trace an error
					// else {
					// 	log.Println("glob no match")
					// }
				}

				// Otherwise, use the catch all (router "fallthrough" equivalent) handler.
				// The application can inspect the map and make a decision on what to do, if anything.
				// This is optional.
				if !handled {
					// It's possible that the S3ObjectRouter wasn't created with NewS3ObjectRouter, so check for this still.
					if handler, ok := r.handlers["_"]; ok {
						r.Tracer.Record("annotation",
							map[string]interface{}{
								"S3Bucket":           record.S3.Bucket.Name,
								"S3ObjectKey":        record.S3.Object.Key,
								"S3Event":            record.EventName,
								"FallthroughHandler": true,
							},
						)

						err = r.Tracer.Capture(ctx, "S3ObjectHandler", func(ctx1 context.Context) error {
							// Tracer is passed in as a dependency so that the handler (user function) can also
							// add annotations, metadata, etc. Whatever the TraceStrategy allows for.
							d.Tracer = &r.Tracer
							return handler.Handler(ctx1, d, &evt)
						})
					}
				}
			}
		}

	}

	return err
}

// Listen will start an S3 event listener that handles incoming object based events (put, delete, etc.)
func (r *S3ObjectRouter) Listen() {
	lambda.Start(r.LambdaHandler)
}

// NewS3ObjectRouter simply returns a new S3ObjectRouter struct and behaves a bit like Router, it even takes an optional rootHandler or "fall through" catch all
func NewS3ObjectRouter(rootHandler ...S3ObjectHandler) *S3ObjectRouter {
	// The catch all is optional, if not provided, an empty handler is still called and it returns nothing.
	// Note: No Event field, so all event sources match for the fall through
	handler := S3ObjectHandler{
		Handler: func(context.Context, *HandlerDependencies, *S3Event) error {
			return nil
		},
	}
	if rootHandler != nil {
		handler = rootHandler[0]
	}
	return &S3ObjectRouter{
		handlers: map[string]S3ObjectHandler{
			"_": handler,
		},
	}
}

// NewS3ObjectRouterForBucket is the same as NewS3ObjectRouter except it's for a specific bucket (you could also set the bucket field after using the other function)
func NewS3ObjectRouterForBucket(bucket string, rootHandler ...S3ObjectHandler) *S3ObjectRouter {
	var r *S3ObjectRouter
	if rootHandler != nil {
		r = NewS3ObjectRouter(rootHandler[0])
	} else {
		r = NewS3ObjectRouter()
	}
	// Just convenience
	r.Bucket = bucket
	return r
}

// Handle will register a handler for a given S3 object event and key name glob match
func (r *S3ObjectRouter) Handle(event string, keyMatch string, handler func(context.Context, *HandlerDependencies, *S3Event) error) {
	if r.handlers == nil {
		r.handlers = make(map[string]S3ObjectHandler)
	}
	r.handlers[keyMatch] = S3ObjectHandler{
		Handler: handler,
		Event:   event,
	}
}

// Created is the same as Handle only the event is already implied. It handles any ObjectCreated event.
func (r *S3ObjectRouter) Created(keyMatch string, handler func(context.Context, *HandlerDependencies, *S3Event) error) {
	r.Handle("s3:ObjectCreated:*", keyMatch, handler)
}

// Put is the same as Handle only the event is already implied.
func (r *S3ObjectRouter) Put(keyMatch string, handler func(context.Context, *HandlerDependencies, *S3Event) error) {
	r.Handle("s3:ObjectCreated:Put", keyMatch, handler)
}

// Post is the same as Handle only the event is already implied.
func (r *S3ObjectRouter) Post(keyMatch string, handler func(context.Context, *HandlerDependencies, *S3Event) error) {
	r.Handle("s3:ObjectCreated:Post", keyMatch, handler)
}

// Copy is the same as Handle only the event is already implied.
func (r *S3ObjectRouter) Copy(keyMatch string, handler func(context.Context, *HandlerDependencies, *S3Event) error) {
	r.Handle("s3:ObjectCreated:Copy", keyMatch, handler)
}

// CompleteMultipartUpload is the same as Handle only the event is already implied.
func (r *S3ObjectRouter) CompleteMultipartUpload(keyMatch string, handler func(context.Context, *HandlerDependencies, *S3Event) error) {
	r.Handle("s3:ObjectCreated:CompleteMultipartUpload", keyMatch, handler)
}

// Removed is the same as Handle only the event is already implied. It handles any ObjectRemoved event.
func (r *S3ObjectRouter) Removed(keyMatch string, handler func(context.Context, *HandlerDependencies, *S3Event) error) {
	r.Handle("s3:ObjectRemoved:*", keyMatch, handler)
}

// Delete is the same as Handle only the event is already implied.
func (r *S3ObjectRouter) Delete(keyMatch string, handler func(context.Context, *HandlerDependencies, *S3Event) error) {
	r.Handle("s3:ObjectRemoved:Delete", keyMatch, handler)
}

// DeleteMarkerCreated is the same as Handle only the event is already implied.
func (r *S3ObjectRouter) DeleteMarkerCreated(keyMatch string, handler func(context.Context, *HandlerDependencies, *S3Event) error) {
	r.Handle("s3:ObjectRemoved:DeleteMarkerCreated", keyMatch, handler)
}

// ReducedRedundancyLostObject is the same as Handle only the event is already implied.
func (r *S3ObjectRouter) ReducedRedundancyLostObject(keyMatch string, handler func(context.Context, *HandlerDependencies, *S3Event) error) {
	r.Handle("s3:ReducedRedundancyLostObject", keyMatch, handler)
}

// TODO: Add helper function get pre-signed upload URL.
// Also look to add custom metadata.
// https://codingbeauty.wordpress.com/2017/04/25/adding-custom-metadata-on-aws-s3-pre-signed-url-generation/
// The S3Event doesn't appear to return any sort of object metadata though.
