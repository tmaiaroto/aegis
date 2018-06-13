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

func TestHandler(t *testing.T) {

	Convey("getType() should return the type of event sent to Lambda", t, func() {
		evtType := getType(map[string]interface{}{"_taskName": "foo"})
		So(evtType, ShouldEqual, "AegisTask")

		evtType = getType(map[string]interface{}{"_rpcName": "foo"})
		So(evtType, ShouldEqual, "AegisRPC")

		evtType = getType(map[string]interface{}{"userPoolId": "123", "triggerSource": "abc"})
		So(evtType, ShouldEqual, "CognitoTrigger")

		evtType = getType(map[string]interface{}{"identityPoolId": "123", "datasetRecords": map[string]interface{}{}})
		So(evtType, ShouldEqual, "CognitoEvent")

		// S3 events have a "Records" object
		evtType = getType(map[string]interface{}{"Records": []interface{}{
			map[string]interface{}{"s3": "..."},
		}})
		So(evtType, ShouldEqual, "S3Event")

		// similar to S3 is SES which also has "Records" but instead of "s3" it has "ses"
		evtType = getType(map[string]interface{}{"Records": []interface{}{
			map[string]interface{}{"ses": "..."},
		}})
		So(evtType, ShouldEqual, "SimpleEmailEvent")

		evtType = getType(map[string]interface{}{"unknown": "event"})
		So(evtType, ShouldEqual, "")
	})
}
