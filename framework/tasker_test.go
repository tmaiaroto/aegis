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

	. "github.com/smartystreets/goconvey/convey"
)

func TestTasker(t *testing.T) {

	fallThroughHandled := false
	testTasker := NewTasker(func(ctx context.Context, d *HandlerDependencies, evt map[string]interface{}) error {
		fallThroughHandled = true
		return nil
	})
	testTasker.Tracer = &NoTraceStrategy{}
	Convey("NewTasker()", t, func() {
		Convey("Should create a new Tasker", func() {
			So(testTasker, ShouldNotBeNil)
		})
	})

	Convey("Should handle specific events with `_taskName`", t, func() {
		handled := false
		testTasker.Handle("foo", func(ctx context.Context, d *HandlerDependencies, evt map[string]interface{}) error {
			handled = true
			return nil
		})

		So(handled, ShouldBeFalse)
		testTasker.LambdaHandler(context.Background(), &HandlerDependencies{}, map[string]interface{}{"_taskName": "foo"})
		So(handled, ShouldBeTrue)
	})

	Convey("Should optionally handle events with `_taskName` using a fall through handler", t, func() {
		So(fallThroughHandled, ShouldBeFalse)
		testTasker.LambdaHandler(context.Background(), &HandlerDependencies{}, map[string]interface{}{"_taskName": "unrouted"})
		So(fallThroughHandled, ShouldBeTrue)
	})

	Convey("Handle() should register a Tasker (scheduled task) handler", t, func() {
		testTasker.Handle("myTask", func(ctx context.Context, d *HandlerDependencies, evt map[string]interface{}) error { return nil })
		So(testTasker.handlers, ShouldNotBeNil)
		So(testTasker.handlers["myTask"], ShouldNotBeNil)

		testNilTasker := &Tasker{}
		testNilTasker.Handle("anotherTask", func(ctx context.Context, d *HandlerDependencies, evt map[string]interface{}) error { return nil })
		So(testNilTasker.handlers, ShouldNotBeNil)
		So(testNilTasker.handlers["anotherTask"], ShouldNotBeNil)
	})
}
