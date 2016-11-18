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
	"io"
	"log"
	"os"
)

type (
	Handler func(*Context, json.RawMessage) (interface{}, error)

	// Context for the AWS Lambda
	Context struct {
		AwsRequestID             string `json:"awsRequestId"`
		FunctionName             string `json:"functionName"`
		FunctionVersion          string `json:"functionVersion"`
		Invokeid                 string `json:"invokeid"`
		IsDefaultFunctionVersion bool   `json:"isDefaultFunctionVersion"`
		LogGroupName             string `json:"logGroupName"`
		LogStreamName            string `json:"logStreamName"`
		MemoryLimitInMB          string `json:"memoryLimitInMB"`
	}

	// Payload sent to the AWS Lambda handler
	Payload struct {
		// custom event fields
		Event json.RawMessage `json:"event"`

		// default context object
		Context *Context `json:"context"`
	}

	// Response is a typical AWS Lambda return format (error, data)
	Response struct {
		// Any errors that occur during processing
		// or are returned by handlers are returned
		Error *string `json:"error"`
		// General purpose output data
		Data interface{} `json:"data"`
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
	}
)

// NewErrorResponse returns the typical AWS Lambda format for a failure
func NewErrorResponse(err error) *Response {
	e := err.Error()
	return &Response{
		Error: &e,
	}
}

// NewResponse returns the typical AWS Lambda format for success
func NewResponse(data interface{}) *Response {
	return &Response{
		Data: data,
	}
}

// NewProxyResponse returns a response in the required format for using AWS API Gateway with a Lambda Proxy
func NewProxyResponse(c string, h map[string]string, b string) *ProxyResponse {
	return &ProxyResponse{
		StatusCode: c,
		Headers:    h,
		Body:       b,
	}
}

// Handle a normal AWS Lambda function
func Handle(handler Handler) {
	RunStream(handler, os.Stdin, os.Stdout, false)
}

// HandleProxy handles an AWS Lambda function as proxy via API Gateway
func HandleProxy(handler Handler) {
	RunStream(handler, os.Stdin, os.Stdout, true)
}

func RunStream(handler Handler, Stdin io.Reader, Stdout io.Writer, proxy bool) {

	stdin := json.NewDecoder(Stdin)
	stdout := json.NewEncoder(Stdout)

	if err := func() (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("panic: %v", e)
			}
		}()
		var payload Payload
		if err := stdin.Decode(&payload); err != nil {
			// Would be nice to use net/http constants for status code here but AWS Lambda Proxy wants a string.
			// return err
			if proxy {
				return stdout.Encode(NewProxyResponse("500", map[string]string{}, ""))
			}
			return err
		}

		// Call the handler.
		data, err := handler(payload.Context, payload.Event)
		if err != nil {
			if proxy {
				return stdout.Encode(NewProxyResponse("500", map[string]string{}, err.Error()))
			}
			return err
		}

		// Remember data is an interface{}, but needs to be a string for ProxyResponse.
		// For a regular Lambda Response it can be a map.
		if proxy {
			return stdout.Encode(NewProxyResponse("500", map[string]string{}, data.(string)))
		}
		return stdout.Encode(NewResponse(data))
	}(); err != nil {
		if encErr := stdout.Encode(NewErrorResponse(err)); encErr != nil {
			// bad times
			log.Println("Failed to encode err response!", encErr.Error())
		}
	}

}
