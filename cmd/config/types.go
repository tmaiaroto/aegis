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

// various types are held here for re-use

package config

// CloudWatchRuleEventPattern defines an event pattern (ultimately sent as JSON)
type CloudWatchRuleEventPattern struct {
	Source     []string               `json:"source"`
	DetailType []string               `json:"detail-type"`
	Detail     map[string]interface{} `json:"detail"`
}

// S3BucketPolicy defines a generic bucket policy
// {
// 	"Version": "2012-10-17",
// 	"Id": "default",
// 	"Statement": [
// 	  {
// 		"Sid": "<optional>",
// 		"Effect": "Allow",
// 		"Principal": {
// 		  "Service": "s3.amazonaws.com"
// 		},
// 		"Action": "lambda:InvokeFunction",
// 		"Resource": "<ArnToYourFunction>",
// 		"Condition": {
// 		  "StringEquals": {
// 			"AWS:SourceAccount": "<YourAccountId>"
// 		  },
// 		  "ArnLike": {
// 			"AWS:SourceArn": "arn:aws:s3:::<YourBucketName>"
// 		  }
// 		}
// 	  }
// 	]
// }
type S3BucketPolicy struct {
	Version   string            `json:"Version"`
	ID        string            `json:"Id"`
	Statement []PolicyStatement `json:"Statement"`
}

// PolicyStatement defines a generic AWS policy statement
type PolicyStatement struct {
	Sid       string             `json:"Sid"`
	Effect    string             `json:"Effect"`
	Principal StatementPrincipal `json:"Principal"`
	Action    interface{}        `json:"Action"`   // can be string or []string
	Resource  interface{}        `json:"Resource"` // can be string or []string
	Condition StatementCondition `json:"Condition"`
}

// StatementPrincipal defines a generic AWS policy statement principal (TODO: see what else there is besides Service in here)
type StatementPrincipal struct {
	Service string `json:"Service"`
}

// StatementCondition defines a generic AWS policy statement condition (TODO: Add more fields as needed)
type StatementCondition struct {
	StringEquals map[string]string `json:"StringEquals"`
}
