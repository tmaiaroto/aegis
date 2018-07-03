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
	"bytes"
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

// SQSRouter struct provides an interface to handle SQS events (routers can be per Domain much like S3ObjectRouter can be per Bucket)
type SQSRouter struct {
	handlers map[string]SQSHandler
	Queue    string
	Tracer   TraceStrategy
}

// SQSHandler handles incoming SQS events. Matching on attribute name and its string value (ie. 1 is "1" and binary values are base64 strings)
type SQSHandler struct {
	Handler           func(context.Context, *HandlerDependencies, *SQSEvent) error
	AttributeName     string
	AttributeStrValue string
}

// LambdaHandler handles SQS events.
func (r *SQSRouter) LambdaHandler(ctx context.Context, d *HandlerDependencies, evt SQSEvent) error {
	// If this Router had a Tracer set for it, replace the default which came from the Aegis interface.
	if r.Tracer != nil {
		d.Tracer = r.Tracer
	}

	// If an incoming event can be matched to this router, but the router has no registered handlers
	// or if one hasn't been added to aegis.Handlers{}.
	if r == nil {
		return errors.New("no handlers registered for SQSRouter")
	}
	var err error
	handled := false

	// There can be multiple
	for _, record := range evt.Records {
		// If there are any handlers registered
		if r.handlers != nil {
			// First, look for queue name match
			queueMatch := false
			if record.EventSourceARN == r.Queue || GetQueueNameFromARN(record.EventSourceARN) == r.Queue {
				queueMatch = true
			}
			// and if the domain matches or if no domain was defined for the SQSRouter (all domains)
			if r.Queue == "" || queueMatch {
				// Each key on handlers map is a string that can be used to glob match
				// If one of them matches, use that
				matchedAttributeName := ""
				matchedAttributeStrValue := ""
				for _, handler := range r.handlers {
					attributeMatch := false
					for attrName, attrVal := range record.MessageAttributes {
						if attrVal.StringValue != nil {
							strVal := *attrVal.StringValue
							if handler.AttributeName == attrName && handler.AttributeStrValue == strVal {
								attributeMatch = true
							}
						}
					}

					if attributeMatch {
						handled = true
						// Trace (default is to use XRay)
						d.Tracer.Record("annotation",
							map[string]interface{}{
								"EventSourceARN":    record.EventSourceARN,
								"EventSource":       record.EventSource,
								"AWSRegion":         record.AWSRegion,
								"AttributeName":     matchedAttributeName,
								"AttributeStrValue": matchedAttributeStrValue,
							},
						)

						err = d.Tracer.Capture(ctx, "SQSHandler", func(ctx1 context.Context) error {
							return handler.Handler(ctx1, d, &evt)
						})
					}
				}

				// Otherwise, use the catch all (router "fallthrough" equivalent) handler.
				// The application can inspect the map and make a decision on what to do, if anything.
				// This is optional.
				if !handled {
					// It's possible that the SQSRouter wasn't created with NewSQSRouter, so check for this still.
					if handler, ok := r.handlers["_"]; ok {
						// Capture the handler (in XRay by default) automatically
						d.Tracer.Record("annotation",
							map[string]interface{}{
								"EventSourceARN":     record.EventSourceARN,
								"EventSource":        record.EventSource,
								"AWSRegion":          record.AWSRegion,
								"AttributeName":      matchedAttributeName,
								"AttributeStrValue":  matchedAttributeStrValue,
								"FallthroughHandler": true,
							},
						)

						err = d.Tracer.Capture(ctx, "SQSHandler", func(ctx1 context.Context) error {
							return handler.Handler(ctx1, d, &evt)
						})
					}
				}
			}
		}

	}

	return err
}

// Listen will start an SES event listener that handles incoming object based events (put, delete, etc.)
func (r *SQSRouter) Listen() {
	lambda.Start(r.LambdaHandler)
}

// NewSQSRouter simply returns a new SQSRouter struct and behaves a bit like Router, it even takes an optional rootHandler or "fall through" catch all
func NewSQSRouter(rootHandler ...func(context.Context, *HandlerDependencies, *SQSEvent) error) *SQSRouter {
	// The catch all is optional, if not provided, an empty handler is still called and it returns nothing.
	handler := SQSHandler{
		Handler: func(context.Context, *HandlerDependencies, *SQSEvent) error {
			return nil
		},
	}
	if len(rootHandler) > 0 {
		handler = SQSHandler{
			Handler: rootHandler[0],
		}
	}
	return &SQSRouter{
		handlers: map[string]SQSHandler{
			"_": handler,
		},
	}
}

// Handle will register a handler for a given attribute name and its string value match (1 would be "1" and binary values would be base64 strings).
// Note that routers can be per queue so this match could match two different queues and therefore be used in two different routers.
func (r *SQSRouter) Handle(attrName string, attrStrValue string, handler func(context.Context, *HandlerDependencies, *SQSEvent) error) {
	if r.handlers == nil {
		r.handlers = make(map[string]SQSHandler)
	}
	var buffer bytes.Buffer
	buffer.WriteString(attrName)
	buffer.WriteString(attrStrValue)
	k := buffer.String()
	buffer.Reset()
	r.handlers[k] = SQSHandler{
		Handler:           handler,
		AttributeName:     attrName,
		AttributeStrValue: attrStrValue,
	}
}

// GetQueueNameFromARN will get the SQS queue name given the ARN string
func GetQueueNameFromARN(arn string) string {
	p := strings.Split(arn, ":")
	return p[len(p)-1]
}

// TODO: Helper perhaps to send messages to a queue?
