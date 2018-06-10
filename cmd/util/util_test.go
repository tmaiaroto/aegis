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

package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUtil(t *testing.T) {
	Convey("GetAccountInfoFromLambdaArn", t, func() {
		Convey("Should return account info from a given Lamba ARN", func() {
			arn := "arn:aws:lambda:us-east-1:1234567890:function:aegis_example:1"
			account, region := GetAccountInfoFromLambdaArn(arn)
			So(account, ShouldEqual, "1234567890")
			So(region, ShouldEqual, "us-east-1")
		})
	})

	Convey("StripLamdaVersionFromArn", t, func() {
		Convey("Should strip the Lambda version number from a given Lambda ARN", func() {
			arn := "arn:aws:lambda:us-east-1:1234567890:function:aegis_example:1"
			expected := "arn:aws:lambda:us-east-1:1234567890:function:aegis_example"
			So(StripLamdaVersionFromArn(arn), ShouldEqual, expected)
		})

		Convey("Should return the versionless ARN if no version was given", func() {
			arn := "arn:aws:lambda:us-east-1:1234567890:function:aegis_example"
			expected := "arn:aws:lambda:us-east-1:1234567890:function:aegis_example"
			So(StripLamdaVersionFromArn(arn), ShouldEqual, expected)
		})
	})
}
