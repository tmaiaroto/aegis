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

package main

import (

	//"encoding/json"
	"bytes"
	"context"
	"errors"
	"log"
	"net/url"

	aegis "github.com/tmaiaroto/aegis/framework"
	"github.com/unrolled/secure"
)

// AegisApp holds a bunch of services that each handler might need
var AegisApp *aegis.Aegis

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
	tasker := aegis.NewTasker(taskerFallThrough)
	tasker.Handle("test", handleTask)

	// Handle incoming emails
	sesReceiver := aegis.NewSESRouter()
	sesReceiver.Handle("*@ses.serifandsemaphore.io", handleEmail)

	// Handle with a URL reqeust path Router
	router := aegis.NewRouter(fallThrough)

	// Aegis can use standard middleware as well.
	// So it's real easy to use something like this secure package.
	// That will help us add some headers for security.
	secureMiddleware := secure.New(secure.Options{
		FrameDeny: true,
	})
	router.UseStandard(secureMiddleware.Handler)

	router.Handle("GET", "/", root)
	router.Handle("GET", "/blah/:thing", somepath, fooMiddleware, barMiddleware)

	router.Handle("POST", "/", postExample)

	// Handle RPCs
	rpcRouter := aegis.NewRPCRouter()
	rpcRouter.Handle("procedure", handleProcedure)

	// Handle S3 objects
	s3Router := aegis.NewS3ObjectRouterForBucket("aegis-incoming")
	// s3Router.Handle("s3:ObjectCreated:Put", "*.png", handleS3Upload)
	// Put() is a shortcut for the above
	s3Router.Put("*.png", handleS3Upload)

	// Blocks. So this function would only be good for handling APIGatewayProxyRequest events
	// router.Listen()
	// Also blocks, but uses reflection to get the event type and then calls the appropriate handler
	// This way, the same Go application can be used to handle multiple events.
	// This is a microservice design consideration. To each their own.
	handlers := aegis.Handlers{
		Router:         router,
		Tasker:         tasker,
		RPCRouter:      rpcRouter,
		S3ObjectRouter: s3Router,
		SESRouter:      sesReceiver,
	}
	// This still works, but it skips service set up.
	// Not using Cognito or any other service in your handlers? Great! Feel free to call this.
	// handlers.Listen()
	//
	// Using the Aegis interface isn't even necessary in this case either. Especially if
	// you don't want any custom logging, etc. The choice is yours.
	// Of course handlers.Listen() is also entirely optional too. Each router/handler has
	// the ability to be used by itself with lambda.Start(). Each router/handler has a
	// LambdaHandler function for that. So, lambda.Start(router.LambdaHandler) for example.
	// That's 3 ways to handle Lambdas.
	// An RPC only Lambda for example may really wish to be thinner and not include
	// all the service config, other handlers, etc.

	AegisApp = aegis.New(handlers)
	AegisApp.Start()
}

func fallThrough(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	res.StatusCode = 404
	return nil
}

func root(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	aegis.Log.Info("logging to CloudWatch")
	log.Println("normal go logging (also goes to cloudwatch)")

	res.JSON(200, map[string]interface{}{"event": req})
	return nil
}

func postExample(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	form, err := req.GetForm()
	if err != nil {
		res.JSON(500, err.Error())
	} else {
		res.JSON(200, form)
	}
	return nil
}

func somepath(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
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

	return errors.New("test error")
}

// Notice the Middleware has a return type. True means go to the next middleware. False
// means to stop right here. If you return false to end the request-response cycle you MUST
// write something back to the client, otherwise it will be left hanging.
func fooMiddleware(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) bool {
	log.Println("Foo!")
	return true
}

func barMiddleware(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) bool {
	log.Println("Bar!")
	res.Body = "bar!"
	return true
}

// Example task handler
func handleTask(ctx context.Context, d *aegis.HandlerDependencies, evt map[string]interface{}) error {
	log.Println("Handling task!", evt)
	return nil
}

// Example task handler catch all
func taskerFallThrough(ctx context.Context, d *aegis.HandlerDependencies, evt map[string]interface{}) error {
	log.Println("Handling task!", evt)
	return nil
}

// Example RPC handler
func handleProcedure(ctx context.Context, d *aegis.HandlerDependencies, evt map[string]interface{}) (map[string]interface{}, error) {
	log.Println("Handling remote procedure!")
	return evt, nil
}

// Example S3 handler
func handleS3Upload(ctx context.Context, d *aegis.HandlerDependencies, evt *aegis.S3Event) error {
	log.Println("Handling S3 upload!")
	log.Println(evt)
	return nil
}

func handleEmail(ctx context.Context, d *aegis.HandlerDependencies, evt *aegis.SimpleEmailEvent) error {
	log.Println("Handling an incoming e-mail!")
	log.Println(evt)
	return nil
}
