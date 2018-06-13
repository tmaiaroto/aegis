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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTracer(t *testing.T) {

	testTracer := &XRayTraceStrategy{}
	Convey("XRayTraceStrategy", t, func() {

		Convey("Record() should set annotations, namespaced annotations, or an error", func() {
			testTracer.Record("annotations", map[string]interface{}{
				"Key":  "value",
				"Key2": "value2",
			})
			So(testTracer.Annotation, ShouldNotBeNil)
			So(len(testTracer.Annotation), ShouldEqual, 2)

			testTracer.Record("metadata", map[string]interface{}{
				"Key": "value",
			})
			So(testTracer.Metadata, ShouldNotBeNil)
			So(len(testTracer.Metadata), ShouldEqual, 1)

			testTracer.Record("namespaceMetadata", map[string]map[string]interface{}{
				"namespace": map[string]interface{}{
					"Key":  "value",
					"Key2": "value2",
					"Key3": "value3",
				},
			})
			So(testTracer.NamespaceMetadata, ShouldNotBeNil)
			So(len(testTracer.NamespaceMetadata), ShouldEqual, 1)
			So(len(testTracer.NamespaceMetadata["namespace"]), ShouldEqual, 3)

			testTracer.Record("error", errors.New("it did not work"))
			So(testTracer.Error, ShouldNotBeNil)
			So(testTracer.Error, ShouldBeError)
		})
	})

	Convey("NoTraceStrategy", t, func() {

		testNoTracer := &NoTraceStrategy{}

		Convey("BeginSegment() should return context", func() {
			ctx := context.Background()
			ctx1, data := testNoTracer.BeginSegment(ctx, "foo")
			So(ctx1, ShouldHaveSameTypeAs, ctx)
			So(data, ShouldBeNil)
		})

		Convey("BeginSubsegment() should return context", func() {
			ctx := context.Background()
			ctx1, data := testNoTracer.BeginSubsegment(ctx, "bar")
			So(ctx1, ShouldHaveSameTypeAs, ctx)
			So(data, ShouldBeNil)
		})
	})
}
