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
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHandlerHelpers(t *testing.T) {
	Convey("GetVariable()", t, func() {
		d := HandlerDependencies{Services: &Services{}}
		testVars := map[string]string{"foo": "bar"}
		d.Services.Variables = testVars

		Convey("Should be able to return a variable value given its key", func() {
			val := d.GetVariable("foo")
			So(val, ShouldEqual, "bar")
		})

		Convey("Should return an OS environment variable fallback", func() {
			os.Setenv("testvar", "notbaragain")
			val := d.GetVariable("testvar")
			So(val, ShouldEqual, "notbaragain")
			os.Setenv("testvar", "")
		})
	})

}
