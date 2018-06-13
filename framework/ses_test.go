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
	"testing"

	events "github.com/aws/aws-lambda-go/events"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSESRouter(t *testing.T) {

	fallThroughHandled := false
	domainSESRouter := NewSESRouterForDomain("example.com", func(ctx context.Context, d *HandlerDependencies, evt *SimpleEmailEvent) error {
		fallThroughHandled = true
		return nil
	})
	domainSESRouter.Tracer = &NoTraceStrategy{}

	fallThroughDomainlessHandled := false
	domainlessSESRouter := NewSESRouter(func(ctx context.Context, d *HandlerDependencies, evt *SimpleEmailEvent) error {
		fallThroughDomainlessHandled = true
		return nil
	})
	domainlessSESRouter.Tracer = &NoTraceStrategy{}

	Convey("NewSESRouterForDomain", t, func() {
		Convey("Should create a new SESRouter for a specific domain", func() {
			So(domainSESRouter, ShouldNotBeNil)
			So(domainSESRouter.Domain, ShouldEqual, "example.com")
		})
	})

	Convey("Should handle `SimpleEmailEvent` events", t, func() {
		handled := false
		domainSESRouter.Handle("user@example.com", func(ctx context.Context, d *HandlerDependencies, evt *SimpleEmailEvent) error {
			handled = true
			return nil
		})

		So(handled, ShouldBeFalse)
		domainSESRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, SimpleEmailEvent{
			Records: []events.SimpleEmailRecord{
				events.SimpleEmailRecord{
					SES: events.SimpleEmailService{
						Mail: events.SimpleEmailMessage{},
						Receipt: events.SimpleEmailReceipt{
							Recipients: []string{"user@example.com"},
						},
					},
				},
			},
		})
		So(handled, ShouldBeTrue)
	})

	Convey("Should optionally handle `SimpleEmailEvent` events using a fall through handler", t, func() {
		So(fallThroughHandled, ShouldBeFalse)
		domainSESRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, SimpleEmailEvent{
			Records: []events.SimpleEmailRecord{
				events.SimpleEmailRecord{
					SES: events.SimpleEmailService{
						Mail: events.SimpleEmailMessage{},
						Receipt: events.SimpleEmailReceipt{
							Recipients: []string{"another@example.com"},
						},
					},
				},
			},
		})
		So(fallThroughHandled, ShouldBeTrue)

		So(fallThroughDomainlessHandled, ShouldBeFalse)
		domainlessSESRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, SimpleEmailEvent{
			Records: []events.SimpleEmailRecord{
				events.SimpleEmailRecord{
					SES: events.SimpleEmailService{
						Mail: events.SimpleEmailMessage{},
						Receipt: events.SimpleEmailReceipt{
							Recipients: []string{"another@unhandled.com"},
						},
					},
				},
			},
		})
		So(fallThroughDomainlessHandled, ShouldBeTrue)
	})

	Convey("Handle() should register a Tasker (scheduled task) handler", t, func() {
		domainSESRouter.Handle("foo@bar.com", func(ctx context.Context, d *HandlerDependencies, evt *SimpleEmailEvent) error {
			return nil
		})
		So(domainSESRouter.handlers, ShouldNotBeNil)
		So(domainSESRouter.handlers["foo@bar.com"], ShouldNotBeNil)

		testNilSESRouter := &SESRouter{}
		testNilSESRouter.Handle("someone@somewhere.com", func(ctx context.Context, d *HandlerDependencies, evt *SimpleEmailEvent) error { return nil })
		So(testNilSESRouter.handlers, ShouldNotBeNil)
		So(testNilSESRouter.handlers["someone@somewhere.com"], ShouldNotBeNil)
	})

}
