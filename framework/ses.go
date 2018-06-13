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
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gobwas/glob"
)

// SESRouter struct provides an interface to handle SES events (routers can be per Domain much like S3ObjectRouter can be per Bucket)
type SESRouter struct {
	handlers map[string]SESHandler
	Domain   string
	Tracer   TraceStrategy
}

// SESHandler handles incoming SES e-mail message events
type SESHandler func(context.Context, *HandlerDependencies, *SimpleEmailEvent) error

// LambdaHandler handles SES received e-mail events.
func (r *SESRouter) LambdaHandler(ctx context.Context, d *HandlerDependencies, evt SimpleEmailEvent) error {
	// If an incoming event can be matched to this router, but the router has no registered handlers
	// or if one hasn't been added to aegis.Handlers{}.
	if r == nil {
		return errors.New("no handlers registered for SESRouter")
	}
	var err error
	var g glob.Glob
	handled := false

	// There can apparently be multiple, a router handler is to handle one object/file operation
	for _, record := range evt.Records {
		// If there are any handlers registered
		if r.handlers != nil {
			// Get all the domains involved, and organize the e-mail addresses
			recordDomains := map[string][]string{}
			for _, recipient := range record.SES.Receipt.Recipients {
				components := strings.Split(recipient, "@")
				_, domain := components[0], components[1]
				if recordDomains[domain] == nil {
					recordDomains[domain] = []string{recipient}
				} else {
					recordDomains[domain] = append(recordDomains[domain], recipient)
				}
			}

			// First, look for domain match
			domainMatch := false
			if _, ok := recordDomains[r.Domain]; ok {
				domainMatch = true
			}
			// and if the domain matches or if no domain was defined for the SESRouter (all domains)
			if r.Domain == "" || domainMatch {
				// Each key on handlers map is a string that can be used to glob match
				// If one of them matches, use that
				for globStr, handler := range r.handlers {
					g = glob.MustCompile(globStr)

					recipientMatch := false
					for _, recipient := range record.SES.Receipt.Recipients {
						if g.Match(recipient) {
							recipientMatch = true
						}
					}

					if recipientMatch {
						handled = true
						// Trace (default is to use XRay)
						r.Tracer.Record("annotation",
							map[string]interface{}{
								"Recipients":     record.SES.Receipt.Recipients,
								"InvocationType": record.SES.Receipt.Action.InvocationType,
							},
						)

						err = r.Tracer.Capture(ctx, "SESHandler", func(ctx1 context.Context) error {
							d.Tracer = &r.Tracer
							return handler(ctx1, d, &evt)
						})
					}
					// TODO: think about some verbose setting for the framework.
					// else {
					// 	log.Println("no match")
					// }
				}

				// Otherwise, use the catch all (router "fallthrough" equivalent) handler.
				// The application can inspect the map and make a decision on what to do, if anything.
				// This is optional.
				if !handled {
					// It's possible that the SESRouter wasn't created with NewSESRouter, so check for this still.
					if handler, ok := r.handlers["_"]; ok {
						// Capture the handler (in XRay by default) automatically
						r.Tracer.Record("annotation",
							map[string]interface{}{
								"Recipients":         record.SES.Receipt.Recipients,
								"InvocationType":     record.SES.Receipt.Action.InvocationType,
								"FallthroughHandler": true,
							},
						)

						err = r.Tracer.Capture(ctx, "SESHandler", func(ctx1 context.Context) error {
							d.Tracer = &r.Tracer
							return handler(ctx1, d, &evt)
						})
					}
				}
			}
		}

	}

	return err
}

// Listen will start an SES event listener that handles incoming object based events (put, delete, etc.)
func (r *SESRouter) Listen() {
	lambda.Start(r.LambdaHandler)
}

// NewSESRouter simply returns a new SESRouter struct and behaves a bit like Router, it even takes an optional rootHandler or "fall through" catch all
func NewSESRouter(rootHandler ...SESHandler) *SESRouter {
	// The catch all is optional, if not provided, an empty handler is still called and it returns nothing.
	handler := func(context.Context, *HandlerDependencies, *SimpleEmailEvent) error {
		return nil
	}
	if len(rootHandler) > 0 {
		handler = rootHandler[0]
	}
	return &SESRouter{
		handlers: map[string]SESHandler{
			"_": handler,
		},
	}
}

// NewSESRouterForDomain is the same as NewSESRouter except it's for a specific domain (you could also set the domain field after using the other function)
func NewSESRouterForDomain(domain string, rootHandler ...SESHandler) *SESRouter {
	var r *SESRouter
	if len(rootHandler) > 0 {
		r = NewSESRouter(rootHandler[0])
	} else {
		r = NewSESRouter()
	}
	// Just convenience
	r.Domain = domain
	return r
}

// Handle will register a handler for a given e-mail address match (regex). Note that routers can be per domain so this address match could match
// two different domains and therefore be used in two different routers, ie. `user@domainA.com` `user@domainB.com` matching on addressMatcher value `user@`.
func (r *SESRouter) Handle(addressMatcher string, handler func(context.Context, *HandlerDependencies, *SimpleEmailEvent) error) {
	if r.handlers == nil {
		r.handlers = make(map[string]SESHandler)
	}
	r.handlers[addressMatcher] = handler
}

// TODO: Helper perhaps to send e-mail?
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/ses-example-send-email.html
// Could inject an interface into the handler dependencies to make it easy to send email from any handler
// It's already pretty easy, but it could save a few steps
// An e-mail reader might be nice too. That is a little more involved because
// the handler here will get a message ID, but then to read the message we have to read from S3.
// Potentially also use KMS to derypt. So helpers to do all that would quite convenient.
// Especially because I'd say many people want to handle the event and then actually do
// something with the contents of the email, not just the headers.
