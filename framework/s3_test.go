// Copyright © 2016 Tom Maiaroto <tom@shift8creative.com>
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

func TestS3ObjectRouter(t *testing.T) {

	bucketRouter := NewS3ObjectRouterForBucket("bucket-name")
	Convey("NewS3ObjectRouterForBucket", t, func() {
		Convey("Should create a new S3ObjectRouter for a specific bucket", func() {
			So(bucketRouter, ShouldNotBeNil)
			So(bucketRouter.Bucket, ShouldEqual, "bucket-name")
		})
	})

}
