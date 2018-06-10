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
	"bytes"
	"regexp"
)

// Helper functions for deployment

// StripLamdaVersionFromArn will remove the :123 version number from a given Lambda ARN, which indicates to use the latest version when used in AWS
func StripLamdaVersionFromArn(lambdaArn string) string {
	// arn:aws:lambda:us-east-1:1234567890:function:aegis_example:1
	r, _ := regexp.Compile("arn:aws:lambda:(.+):([0-9]+):function:([A-z0-9\\-\\_]+)($|:[0-9]+)")
	matches := r.FindStringSubmatch(lambdaArn)
	accountID := ""
	region := ""
	functionName := ""
	if len(matches) == 5 {
		region = matches[1]
		accountID = matches[2]
		functionName = matches[3]
		// functionVersion = matches[4]
	}

	var buffer bytes.Buffer
	buffer.WriteString("arn:aws:lambda:")
	buffer.WriteString(region)
	buffer.WriteString(":")
	buffer.WriteString(accountID)
	buffer.WriteString(":function:")
	buffer.WriteString(functionName)
	arn := buffer.String()
	buffer.Reset()

	return arn
}

// GetAccountInfoFromLambdaArn will extract the account ID and region from a given ARN
func GetAccountInfoFromLambdaArn(lambdaArn string) (string, string) {
	r, _ := regexp.Compile("arn:aws:lambda:(.+):([0-9]+):function")
	matches := r.FindStringSubmatch(lambdaArn)
	accountID := ""
	region := ""
	if len(matches) == 3 {
		region = matches[1]
		accountID = matches[2]
	}

	return accountID, region
}
