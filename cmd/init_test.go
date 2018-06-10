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

func TestInitCmd(t *testing.T) {
	Convey("Should copy a boilerplate yaml config file", t, func() {
		testConfigFilePath := "./aegis-config.yaml.test"
		err := copyConfig(testConfigFilePath)
		So(err, ShouldBeNil)
		// file should exist
		_, err = os.Stat(testConfigFilePath)
		So(err, ShouldBeNil)

		Convey("Should not overwrite an existing file by the same name", func() {
			err = copyConfig(testConfigFilePath)
			So(err, ShouldNotBeNil)
		})

		// cleanup
		_ = os.Remove(testConfigFilePath)
	})

	Convey("Should copy a boilerplate .go source file", t, func() {
		testSrcFilePath := "./aegis-main.go.test"
		err := copySrc(testSrcFilePath)
		So(err, ShouldBeNil)
		// file should exist
		_, err = os.Stat(testSrcFilePath)
		So(err, ShouldBeNil)

		Convey("Should not overwrite an existing file by the same name", func() {
			err = copySrc(testSrcFilePath)
			So(err, ShouldNotBeNil)
		})

		// cleanup
		_ = os.Remove(testSrcFilePath)
	})
}
