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
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-lambda-go/events"
)

// The types here are aliasing AWS Lambda's events package types. This is so Aegis can add some additional functionality.
// @see helpers.go
type (
	// APIGatewayProxyResponse alias for APIGatewayProxyResponse events, additional functionality added by helpers.go
	APIGatewayProxyResponse events.APIGatewayProxyResponse

	// APIGatewayProxyRequest alias for incoming APIGatewayProxyRequest events
	APIGatewayProxyRequest events.APIGatewayProxyRequest

	// APIGatewayProxyRequestContext alias for APIGatewayProxyRequestContext
	APIGatewayProxyRequestContext events.APIGatewayProxyRequestContext

	// CloudWatchEvent alias for CloudWatchEvent events
	CloudWatchEvent events.CloudWatchEvent
)

// Log uses Logrus for logging and will hook to CloudWatch...But could also be used to hook to other centralized logging services.
var Log = logrus.New()
