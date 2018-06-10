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

package cmd

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeployCmd(t *testing.T) {
	Convey("compress", t, func() {
		Convey("Should compress a Lambda function zip file and return the file path", func() {
			testZipFileName := "./aegis_function.test.zip"
			filePath := compress(testZipFileName)
			// make the "aegis_app" file so it exists to zip
			_, _ = os.Create(aegisAppName)
			So(filePath, ShouldContainSubstring, "aegis_function.test.zip")

			// cleanup
			_ = os.Remove(aegisAppName)
			_ = os.Remove(testZipFileName)
		})
	})

	Convey("getExecPath", t, func() {
		Convey("Should return a given executable file's path", func() {
			So(getExecPath("go"), ShouldNotBeEmpty)
		})
	})
}
