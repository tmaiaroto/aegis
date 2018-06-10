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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTasker(t *testing.T) {

	testTasker := NewTasker()
	Convey("NewTasker", t, func() {
		Convey("Should create a new Tasker", func() {
			So(testTasker, ShouldNotBeNil)
		})
	})

	// TODO: Figure out how to run tests with XRay
	// testTaskerVal := 0
	// testCtx := context.Background()
	// testHandler := func(testCtx context.Context, evt map[string]interface{}) error {
	// 	testTaskerVal = 1
	// 	return nil
	// }
	// testEvt := map[string]interface{}{
	// 	"_taskName": "test task",
	// 	"foo":       "bar",
	// }

	// os.Setenv("AWS_XRAY_CONTEXT_MISSING", "LOG_ERROR")
	// testTasker.Handle("test task", testHandler)
	// err := testTasker.LambdaHandler(testCtx, testEvt)

	// Convey("LambdaHandler", t, func() {
	// 	Convey("Should handle proper event", func() {
	// 		So(err, ShouldBeNil)
	// 		So(testTaskerVal, ShouldEqual, 1)
	// 	})
	// })
}
