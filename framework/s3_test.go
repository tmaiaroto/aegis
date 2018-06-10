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

func TestS3ObjectRouter(t *testing.T) {

	bucketRouter := NewS3ObjectRouterForBucket("bucket-name")
	// Allows tests to run without needing AWS credentials
	// ...well, as best they can, remember that a lot of functionality here depends on AWS
	// Note: If a Tracer is not set, there will be a panic
	bucketRouter.Tracer = &NoTraceStrategy{}

	// Fake event and record
	evtSrc := "s3:ObjectCreated:Put"
	s3Key := "foobar.png"
	record := events.S3EventRecord{
		EventSource: evtSrc,
		S3: events.S3Entity{
			Bucket: events.S3Bucket{
				Name: "bucket-name",
			},
			Object: events.S3Object{
				Key: s3Key,
			},
		},
	}
	putEvt := S3Event{
		Records: []events.S3EventRecord{record},
	}

	Convey("NewS3ObjectRouterForBucket()", t, func() {

		Convey("Should create a new S3ObjectRouter for a specific bucket", func() {
			So(bucketRouter, ShouldNotBeNil)
			So(bucketRouter.Bucket, ShouldEqual, "bucket-name")
		})

		Convey("Should handle any S3 object event with a fall through handler", func() {
			handled := false
			bucketRouterWithFallthrough := NewS3ObjectRouterForBucket("bucket-name", S3ObjectHandler{Handler: func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
				So(evt.Records[0].EventSource, ShouldEqual, evtSrc)
				So(evt.Records[0].S3.Object.Key, ShouldEqual, s3Key)
				handled = true
				return nil
			}})
			// Note: If a Tracer is not set, there will be a panic
			bucketRouterWithFallthrough.Tracer = &NoTraceStrategy{}
			// Handle fake event
			bucketRouterWithFallthrough.LambdaHandler(context.Background(), &HandlerDependencies{}, putEvt)
			So(handled, ShouldBeTrue)
		})

		Convey("Created() Should handle a `s3:ObjectCreated` event of any type", func() {
			handled := false
			bucketRouter.Created("*.png", func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
				So(evt.Records[0].EventSource, ShouldEqual, evtSrc)
				So(evt.Records[0].S3.Object.Key, ShouldEqual, s3Key)
				handled = true
				return nil
			})
			// Handle fake event
			bucketRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, putEvt)
			So(handled, ShouldBeTrue)
		})

		Convey("Put()", func() {
			Convey("Should handle a `s3:ObjectCreated:Put` event", func() {
				handled := false
				bucketRouter.Put("*.png", func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
					So(evt.Records[0].EventSource, ShouldEqual, evtSrc)
					So(evt.Records[0].S3.Object.Key, ShouldEqual, s3Key)
					handled = true
					return nil
				})
				// Handle fake event
				bucketRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, putEvt)
				So(handled, ShouldBeTrue)
			})

			Convey("Should NOT handle a `s3:ObjectRemoved:Delete` event", func() {
				deleteEvt := putEvt
				deleteEvt.Records[0].EventSource = "s3:ObjectRemoved:Delete"
				handled := false
				bucketRouter.Put("*.png", func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
					handled = true
					return nil
				})
				// Handle fake event
				bucketRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, deleteEvt)
				So(handled, ShouldBeFalse)
			})
		})

		Convey("Post() Should handle a `s3:ObjectCreated:Post` event", func() {
			postEvt := putEvt
			postEvt.Records[0].EventSource = "s3:ObjectCreated:Post"
			handled := false
			bucketRouter.Post("*.png", func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
				handled = true
				return nil
			})
			// Handle fake event
			bucketRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, postEvt)
			So(handled, ShouldBeTrue)
		})

		Convey("Copy() Should handle a `s3:ObjectCreated:Copy` event", func() {
			copyEvt := putEvt
			copyEvt.Records[0].EventSource = "s3:ObjectCreated:Copy"
			handled := false
			bucketRouter.Copy("*.png", func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
				handled = true
				return nil
			})
			// Handle fake event
			bucketRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, copyEvt)
			So(handled, ShouldBeTrue)
		})

		Convey("CompleteMultipartUpload() Should handle a `s3:ObjectCreated:CompleteMultipartUpload` event", func() {
			mEvt := putEvt
			mEvt.Records[0].EventSource = "s3:ObjectCreated:CompleteMultipartUpload"
			handled := false
			bucketRouter.CompleteMultipartUpload("*.png", func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
				handled = true
				return nil
			})
			// Handle fake event
			bucketRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, mEvt)
			So(handled, ShouldBeTrue)
		})

		Convey("Removed() Should handle a `s3:ObjectRemoved` event of any type", func() {
			deleteEvt := putEvt
			deleteEvt.Records[0].EventSource = "s3:ObjectRemoved:Delete"
			handled := false
			bucketRouter.Removed("*.png", func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
				handled = true
				return nil
			})
			// Handle fake event
			bucketRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, deleteEvt)
			So(handled, ShouldBeTrue)
		})

		Convey("Delete()", func() {
			Convey("Should handle a `s3:ObjectRemoved:Delete` event", func() {
				deleteEvt := putEvt
				deleteEvt.Records[0].EventSource = "s3:ObjectRemoved:Delete"
				handled := false
				bucketRouter.Delete("*.png", func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
					handled = true
					return nil
				})
				// Handle fake event
				bucketRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, deleteEvt)
				So(handled, ShouldBeTrue)
			})

			Convey("Should NOT handle a `s3:ObjectRemoved:DeleteMarkerCreated` event", func() {
				deleteEvt := putEvt
				deleteEvt.Records[0].EventSource = "s3:ObjectRemoved:DeleteMarkerCreated"
				handled := false
				bucketRouter.Delete("*.png", func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
					handled = true
					return nil
				})
				// Handle fake event
				bucketRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, deleteEvt)
				So(handled, ShouldBeFalse)
			})
		})

		Convey("DeleteMarkerCreated() Should handle a `s3:ObjectRemoved:DeleteMarkerCreated` event", func() {
			deleteEvt := putEvt
			deleteEvt.Records[0].EventSource = "s3:ObjectRemoved:DeleteMarkerCreated"
			handled := false
			bucketRouter.DeleteMarkerCreated("*.png", func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
				handled = true
				return nil
			})
			// Handle fake event
			bucketRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, deleteEvt)
			So(handled, ShouldBeTrue)
		})

		Convey("ReducedRedundancyLostObject() Should handle a `s3:ReducedRedundancyLostObject` event", func() {
			rrEvt := putEvt
			rrEvt.Records[0].EventSource = "s3:ReducedRedundancyLostObject"
			handled := false
			bucketRouter.ReducedRedundancyLostObject("*.png", func(ctx context.Context, d *HandlerDependencies, evt *S3Event) error {
				handled = true
				return nil
			})
			// Handle fake event
			bucketRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, rrEvt)
			So(handled, ShouldBeTrue)
		})

	})

}
