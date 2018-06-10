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

func TestCognitoTriggerTypes(t *testing.T) {

	Convey("GetCognitoTriggerType()", t, func() {

		Convey("Should be return a struct name for a Cognito trigger event", func() {
			name := GetCognitoTriggerType(map[string]interface{}{"triggerSource": "not_handled"})
			So(name, ShouldEqual, "")

			// CognitoTriggerPreSignup
			name = GetCognitoTriggerType(map[string]interface{}{"triggerSource": "PreSignUp_AdminCreateUser"})
			So(name, ShouldEqual, "CognitoTriggerPreSignup")

			name = GetCognitoTriggerType(map[string]interface{}{"triggerSource": "PreSignUp_SignUp"})
			So(name, ShouldEqual, "CognitoTriggerPreSignup")

			// CognitoTriggerPostConfirmation
			name = GetCognitoTriggerType(map[string]interface{}{"triggerSource": "PostConfirmation_ConfirmSignUp"})
			So(name, ShouldEqual, "CognitoTriggerPostConfirmation")

			name = GetCognitoTriggerType(map[string]interface{}{"triggerSource": "PostConfirmation_ConfirmForgotPassword"})
			So(name, ShouldEqual, "CognitoTriggerPostConfirmation")

			// CognitoTriggerCustomMessage
			name = GetCognitoTriggerType(map[string]interface{}{"triggerSource": "CustomMessage_ResendCode"})
			So(name, ShouldEqual, "CognitoTriggerCustomMessage")

			// CognitoTriggerPostAuthentication
			name = GetCognitoTriggerType(map[string]interface{}{"triggerSource": "PostAuthentication_Authentication"})
			So(name, ShouldEqual, "CognitoTriggerPostAuthentication")

			// CognitoTriggerPreAuthentication
			name = GetCognitoTriggerType(map[string]interface{}{"triggerSource": "PreAuthentication_Authentication"})
			So(name, ShouldEqual, "CognitoTriggerPreAuthentication")

			// CognitoTriggerTokenGeneration
			name = GetCognitoTriggerType(map[string]interface{}{"triggerSource": "TokenGeneration_HostedAuth"})
			So(name, ShouldEqual, "CognitoTriggerTokenGeneration")

		})
	})

}
