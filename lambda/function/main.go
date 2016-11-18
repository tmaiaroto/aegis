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

package main

import (
	"encoding/json"
	"github.com/tmaiaroto/aegis/lambda"
)

func main() {
	lambda.HandleProxy(func(context *lambda.Context, eventJSON json.RawMessage) (interface{}, error) {
		// Just take the event message passed to this Lambda function and return it.
		var v map[string]interface{}
		if err := json.Unmarshal(eventJSON, &v); err != nil {
			return nil, err
		}
		vStr, _ := json.Marshal(v)

		// Note: Return data as a string.
		// An AWS API Gateway resource using Lamba Proxy needs the Lambda to return a response
		// in a very specific format.
		// This returned string value will be used as part of a JSON response as value for a "body" key.
		return string(vStr), nil
	})
}
