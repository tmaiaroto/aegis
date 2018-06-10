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

// These types are not currently in the aws-lambda-go package (though there's a pull request for partial support).
// The request and response structure depends on the trigger.
// https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-pools-lambda-trigger-syntax-shared.html

// CognitoTriggerCallerContext contains information about the caller (should be the same for all triggers)
type CognitoTriggerCallerContext struct {
	AWSSDKVersion string `json:"awsSdkVersion"`
	ClientID      string `json:"clientId"`
}

// CognitoTriggerCommon contains common data from events sent by AWS Cognito (should be the same for all triggers)
type CognitoTriggerCommon struct {
	Version       string                      `json:"version"`
	TriggerSource string                      `json:"triggerSource"`
	Region        string                      `json:"region"`
	UserPoolID    string                      `json:"userPoolId"`
	CallerContext CognitoTriggerCallerContext `json:"callerContext"`
	UserName      string                      `json:"userName"`
}

// CognitoTriggerPreSignup is invoked when a user submits their information to sign up, allowing you to perform
// custom validation to accept or deny the sign up request.
// triggerSource: PreSignUp_AdminCreateUser, PreSignUp_SignUp
type CognitoTriggerPreSignup struct {
	CognitoTriggerCommon
	Request struct {
		UserAttributes map[string]interface{} `json:"userAttributes"`
		ValidationData map[string]interface{} `json:"validationData"`
	} `json:"request"`
	Response struct {
		AutoConfirmUser bool `json:"autoConfirmUser"`
		AutoVerifyEmail bool `json:"autoVerifyEmail"`
		AutoVerifyPhone bool `json:"autoVerifyPhone"`
	} `json:"response"`
}

// CognitoTriggerPostConfirmation is invoked after a user is confirmed, allowing you to send custom messages
// or to add custom logic, for example for analytics.
// triggerSource: PostConfirmation_ConfirmSignUp, PostConfirmation_ConfirmForgotPassword
type CognitoTriggerPostConfirmation struct {
	CognitoTriggerCommon
	Request struct {
		UserAttributes map[string]interface{} `json:"userAttributes"`
	} `json:"request"`
	Response map[string]interface{} `json:"response"`
}

// CognitoTriggerCustomMessage is invoked before a verification or MFA message is sent, allowing you to
// customize the message dynamically. Note that static custom messages can be edited on the Verifications panel.
// triggerSource: CustomMessage_ResendCode
type CognitoTriggerCustomMessage struct {
	CognitoTriggerCommon
	Request struct {
		UserAttributes    map[string]interface{} `json:"userAttributes"`
		CodeParameter     string                 `json:"codeParameter"`
		UsernameParameter string                 `json:"usernameParameter"`
	} `json:"request"`
	Response struct {
		SMSMessage   string `json:"smsMessage"`
		EmailMessage string `json:"emailMessage"`
		EmailSubject string `json:"emailSubject"`
	} `json:"response"`
}

// CognitoTriggerPostAuthentication is invoked after a user is authenticated, allowing you to add custom logic,
// for example for analytics.
// triggerSource: PostAuthentication_Authentication
type CognitoTriggerPostAuthentication struct {
	CognitoTriggerCommon
	Request struct {
		UserAttributes map[string]interface{} `json:"userAttributes"`
		NewDeviceUsed  bool                   `json:"newDeviceUsed"`
	} `json:"request"`
	Response map[string]interface{} `json:"response"`
}

// CognitoTriggerPreAuthentication is invoked when a user submits their information to be authenticated, allowing
// you to perform custom validations to accept or deny the sign in request.
// triggerSource: PreAuthentication_Authentication
type CognitoTriggerPreAuthentication struct {
	CognitoTriggerCommon
	Request struct {
		UserAttributes map[string]interface{} `json:"userAttributes"`
		ValidationData map[string]interface{} `json:"validationData"`
	} `json:"request"`
	Response map[string]interface{} `json:"response"`
}

// CognitoTriggerTokenGeneration is invoked before the token generation, allowing you to customize the claims
// in the identity token.
// triggerSource: TokenGeneration_HostedAuth
type CognitoTriggerTokenGeneration struct {
	CognitoTriggerCommon
	Request struct {
		UserAttributes     map[string]interface{} `json:"userAttributes"`
		GroupConfiguration map[string]interface{} `json:"groupConfiguration"`
	} `json:"request"`
	Response map[string]interface{} `json:"response"` // default is map[claimsOverrideDetails: <nil>]
}

// GetCognitoTriggerType returns the name of the struct for the Cognito trigger
func GetCognitoTriggerType(evt map[string]interface{}) string {
	// triggerSource key will have the Cognito trigger type, ie. PreSignUp_SignUp
	switch evt["triggerSource"].(string) {
	case "PreSignUp_AdminCreateUser", "PreSignUp_SignUp":
		return "CognitoTriggerPreSignup"
	case "PostConfirmation_ConfirmSignUp", "PostConfirmation_ConfirmForgotPassword":
		return "CognitoTriggerPostConfirmation"
	case "CustomMessage_ResendCode":
		return "CognitoTriggerCustomMessage"
	case "PostAuthentication_Authentication":
		return "CognitoTriggerPostAuthentication"
	case "PreAuthentication_Authentication":
		return "CognitoTriggerPreAuthentication"
	case "TokenGeneration_HostedAuth":
		return "CognitoTriggerTokenGeneration"
	}
	return ""
}
