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
	//"encoding/json"
	"bytes"
	"context"
	"log"
	"net/url"

	aegis "github.com/tmaiaroto/aegis/framework"
)

func main() {
	// Enable line numbers in logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Handle the Lambda Proxy directly
	// aegis.HandleProxy(func(ctx *aegis.Context, evt *events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {

	// 	event, err := json.Marshal(evt)
	// 	if err != nil {
	// 		return aegis.NewProxyResponse(500, map[string]string{}, "", err)
	// 	}

	// 	return aegis.NewProxyResponse(200, map[string]string{}, string(event), nil)
	// })

	// Handle tasks
	tasker := aegis.NewTasker()
	tasker.Handle("somefile.json", handleTask)

	// Handle with a URL reqeust path Router
	router := aegis.NewRouter(fallThrough)

	router.Handle("GET", "/", root)
	router.Handle("GET", "/blah/:thing", somepath, fooMiddleware, barMiddleware)

	router.Handle("POST", "/", postExample)

	// Blocks. So this function would only be good for handling APIGatewayProxyRequest events
	// router.Listen()
	// Also blocks, but uses reflection to get the event type and then calls the appropriate handler
	// This way, the same Go application can be used to handle multiple events.
	// This is a microservice design consideration. To each their own.
	handlers := aegis.Handlers{
		Router: router,
		// Tasker...
	}
	handlers.Listen()
}

func fallThrough(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) {
	res.StatusCode = 404
}

func root(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) {
	aegis.Log.Info("logging to CloudWatch")
	log.Println("normal go logging (also goes to cloudwatch)")

	res.JSON(200, map[string]interface{}{"event": req})
}

func postExample(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) {
	form, err := req.GetForm()
	if err != nil {
		res.JSON(500, err.Error())
	} else {
		res.JSON(200, form)
	}
}

func somepath(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) {
	// log.Println(params) <-- these will be path params...
	// so in this roue definition it means `thing` will be a key.
	// get with: params.Get("thing")

	// concat
	var buffer bytes.Buffer
	buffer.WriteString(res.Body)
	buffer.WriteString(" PLUS MORE!")
	newBody := buffer.String()
	buffer.Reset()
	res.Body = newBody

	// res.Body = "body for /blah/blah"
	// Usually url.Values are used for querysring parameters so Encode() will return a querystring.
	// But in this case...it's used for path params. Regardless, this encode works and makes it easy
	// to see what the path params were.
	res.Body = params.Encode()

	res.Headers = map[string]string{"Content-Type": "application/json"}

	// TODO: Make shortcut/helper functions like res.JSON()
	// which will automatically set content-type header to application/json and take a body input that it will also set.
	// another might be: res.Success(body)
	// and res.Fail(body) ... setting status code 500, etc.
	// (especially because we think of status code as integer, but Lambda Proxy wants a string - helpers are nice)
}

// Notice the Middleware has a return type. True means go to the next middleware. False
// means to stop right here. If you return false to end the request-response cycle you MUST
// write something back to the client, otherwise it will be left hanging.
func fooMiddleware(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) bool {
	log.Println("Foo!")
	return true
}

func barMiddleware(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) bool {
	log.Println("Bar!")
	res.Body = "bar!"
	return true
}

// Example task handler
func handleTask(ctx context.Context, evt *aegis.CloudWatchEvent) {
	log.Println("Handling task!", evt)
}
