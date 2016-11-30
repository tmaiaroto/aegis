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

// Originally from: github.com/jasonmoo/lambda_proc

package lambda

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/tmaiaroto/logrus-cloudwatchlogs"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type (
	// Handler returns int (HTTP status for statusCode), map[string]string (headers), string (body), error (error, which gets placed into the body)
	Handler func(*Context, *Event) *ProxyResponse

	// Context for the AWS Lambda
	Context struct {
		AwsRequestID                   string `json:"awsRequestId"`
		CallbackWaitsForEmptyEventLoop bool   `json:"callbackWaitsForEmptyEventLoop"`
		FunctionName                   string `json:"functionName"`
		FunctionVersion                string `json:"functionVersion"`
		InvokedFunctionArn             string `json:"invokedFunctionArn"`
		Invokeid                       string `json:"invokeid"`
		IsDefaultFunctionVersion       bool   `json:"isDefaultFunctionVersion"`
		LogGroupName                   string `json:"logGroupName"`
		LogStreamName                  string `json:"logStreamName"`
		MemoryLimitInMB                string `json:"memoryLimitInMB"`
	}

	// Event from the API Gateway passed to the AWS Lambda
	Event struct {
		Body            interface{}       `json:"body"`
		Headers         map[string]string `json:"headers"`
		HttpMethod      string            `json:"httpMethod"`
		IsBase64Encoded bool              `json:"isBase64Encoded"`
		Path            string            `json:"path"`
		// Will be {"proxy": "path/parts"} if set.
		// Almost redundant in this case with Path because the API Gateway has this catch all proxy path.
		PathParameters        map[string]string `json:"pathParameters"`
		QueryStringParameters map[string]string `json:"queryStringParameters"`
		RequestContext        RequestContext    `json:"requestContext"`
		// Always `/` or `/{proxy+}` in this case
		Resource           string            `json:"resource"`
		StageVariables     map[string]string `json:"stageVariables"`
		HandlerStartHrTime []int64           `json:"handlerStartHrTime"`
		HandlerStartTimeMs int64             `json:"handlerStartTimeMs"`
		HandlerStartTime   int64             `json:"handlerStartTime"`
	}

	// RequestContext for the API Gateway request (different than the Lambda function context itself)
	RequestContext struct {
		AccountID  string   `json:"accountId"`
		ApiID      string   `json:"apiId"`
		HttpMethod string   `json:"httpMethod"`
		Identity   Identity `json:"identity"`
		RequestId  string   `json:"requestId"`
		ResourceId string   `json:"resourceId"`
		// Always `/` or `/{proxy+}` in this case
		ResourcePath string `json:"resourcePath"`
		Stage        string `json:"stage"`
	}

	// Identity for the API Gateway request
	Identity struct {
		AccessKey                     string `json:"accessKey"`
		AccountId                     string `json:"accountId"`
		ApiKey                        string `json:"apiKey"`
		Caller                        string `json:"caller"`
		CognitoAuthenticationProvider string `json:"cognitoAuthenticationProvider"`
		CognitoAuthenticationType     string `json:"cognitoAuthenticationType"`
		CognitoIdentityId             string `json:"cognitoIdentityId"`
		CognitoIdentityPoolId         string `json:"cognitoIdentityPoolId"`
		SourceIp                      string `json:"sourceIp"`
		User                          string `json:"user"`
		UserAgent                     string `json:"userAgent"`
		UserArn                       string `json:"userArn"`
	}

	// Payload sent to the AWS Lambda handler
	Payload struct {
		// custom event fields
		Event *Event `json:"event"`
		// default context object
		Context *Context `json:"context"`
	}

	// ProxyResponse needs to be a specific format
	// {
	//   "statusCode": "200",
	//   "headers": {
	//     "Content-Type": "application/json"
	//   },
	//   "body": "{\"key1\":\"value1\",\"key2\":\"value2\",\"key3\":\"value3\"}"
	// }
	ProxyResponse struct {
		StatusCode string            `json:"statusCode"`
		Headers    map[string]string `json:"headers"`
		Body       string            `json:"body"`
		err        error             `json:"-"`
	}
)

// Log uses Logrus for logging and will hook to CloudWatch...But could also be used to hook to other centralized logging services.
var Log = logrus.New()

// NewProxyResponse returns a response in the required format for using AWS API Gateway with a Lambda Proxy
func NewProxyResponse(c int, h map[string]string, b string, e error) *ProxyResponse {
	status := strconv.Itoa(c)
	// If there's an error, use that as the body if body is empty.
	if e != nil && b == "" {
		b = e.Error()
	}
	return &ProxyResponse{
		StatusCode: status,
		Headers:    h,
		Body:       b,
		err:        e,
	}
}

// RunStream will take the input passed to the Lambda (wrapper script) from stdio, call the handler (user Go code), and then pipe back a response (suitable for Lambda Proxy)
func RunStream(handler Handler, Stdin io.Reader, Stdout io.Writer) {

	stdin := json.NewDecoder(Stdin)
	stdout := json.NewEncoder(Stdout)

	if err := func() (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("panic: %v", e)
			}
		}()
		var payload Payload
		// If there's a problem with the data coming in...
		if err := stdin.Decode(&payload); err != nil {
			return stdout.Encode(NewProxyResponse(http.StatusInternalServerError, map[string]string{}, "", err))
		}

		// Call the handler.
		// status, headers, body, err := handler(payload.Context, payload.Event)

		// Set up logging through logrus since normal logging via stdout won't work.
		cfg := aws.NewConfig()
		hook, err := logrus_cloudwatchlogs.NewHook(payload.Context.LogGroupName, payload.Context.LogStreamName, cfg)
		if err != nil {
			log.Println("Error setting up logrus hook for CloudWatch")
			log.Fatal(err)
		}
		Log.Hooks.Add(hook)
		Log.Out = ioutil.Discard
		Log.Formatter = logrus_cloudwatchlogs.NewProdFormatter()

		// nanosecond format the millisconds for now
		payload.Event.HandlerStartTime = payload.Event.HandlerStartTimeMs * 1000000
		resp := handler(payload.Context, payload.Event)

		if err != nil {
			// If thre's an error, the statusCode has to be in the 500's.
			// If it isn't, use generic 500.
			if resp.StatusCode == "" {
				resp.StatusCode = strconv.Itoa(http.StatusInternalServerError)
			}
			// return stdout.Encode(NewProxyResponse(status, headers, body, err))
			return stdout.Encode(resp)
		}

		// return stdout.Encode(NewProxyResponse(status, headers, body, nil))
		return stdout.Encode(resp)
	}(); err != nil {
		if encErr := stdout.Encode(NewProxyResponse(http.StatusInternalServerError, map[string]string{}, "", err)); encErr != nil {
			// bad times
			log.Println("Failed to encode err response!", encErr.Error())
		}
	}

}

// HandleProxy handles an AWS Lambda function as proxy via API Gateway directly
func HandleProxy(handler Handler) {
	RunStream(handler, os.Stdin, os.Stdout)
}
