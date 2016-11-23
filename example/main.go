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
	//"log"
	//"net/http"
)

func main() {

	// Handle the Lambda Proxy directly
	lambda.HandleProxy(func(ctx *lambda.Context, evt *lambda.Event) *lambda.ProxyResponse {

		event, err := json.Marshal(evt)
		if err != nil {
			return lambda.NewProxyResponse(500, map[string]string{}, "", err)
		}

		return lambda.NewProxyResponse(200, map[string]string{}, string(event), nil)
	})

	// Handle with a URL reqeust path Router
	// TODO
	//lambda.NewRouter()

}
