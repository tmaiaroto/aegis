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
	"github.com/aws/aws-sdk-go/aws"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSQSRouter(t *testing.T) {

	sqsRouter := NewSQSRouterForQueue("queue-name")
	// Allows tests to run without needing AWS credentials
	// ...well, as best they can, remember that a lot of functionality here depends on AWS
	// Note: If a Tracer is not set, there will be a panic
	sqsRouter.Tracer = NoTraceStrategy{}

	// Fake event and record
	evtSrc := "aws:sqs"
	record := events.SQSMessage{
		EventSourceARN: "arn:aws:sqs:us-east-1:112522220030:queue-name",
		EventSource:    evtSrc,
		MessageAttributes: map[string]events.SQSMessageAttribute{
			"attr1": events.SQSMessageAttribute{
				StringValue: aws.String("someval"),
			},
		},
	}
	sqsEvt := SQSEvent{
		Records: []events.SQSMessage{record},
	}

	Convey("NewSQSRouterForQueue()", t, func() {

		Convey("Should create a new SQSRouter for a specific queue", func() {
			So(sqsRouter, ShouldNotBeNil)
			So(sqsRouter.Queue, ShouldEqual, "queue-name")
		})

		Convey("Should handle any SQS event with a fall through handler", func() {
			handled := false
			sqsRouterWithFallthrough := NewSQSRouterForQueue("queue-name", func(ctx context.Context, d *HandlerDependencies, evt *SQSEvent) error {
				So(evt.Records[0].EventSource, ShouldEqual, evtSrc)
				handled = true
				return nil
			})
			// Note: If a Tracer is not set, there will be a panic
			sqsRouterWithFallthrough.Tracer = NoTraceStrategy{}
			// Handle fake event
			sqsRouterWithFallthrough.LambdaHandler(context.Background(), &HandlerDependencies{}, sqsEvt)
			So(handled, ShouldBeTrue)
		})

		Convey("Should handle an SQS event of based on attribute name/value match", func() {
			handled := false
			sqsRouter.Handle("attr1", "someval", func(ctx context.Context, d *HandlerDependencies, evt *SQSEvent) error {
				So(evt.Records[0].EventSource, ShouldEqual, evtSrc)
				handled = true
				return nil
			})
			// Handle fake event
			sqsRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, sqsEvt)
			So(handled, ShouldBeTrue)
		})

	})

}
