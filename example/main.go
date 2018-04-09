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
	"errors"
	"log"
	"net/url"

	aegis "github.com/tmaiaroto/aegis/framework"
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

	// Handle with a URL reqeust path Router
	router := aegis.NewRouter(fallThrough)

	router.Handle("GET", "/", root)
	router.Handle("GET", "/blah/:thing", somepath, fooMiddleware, barMiddleware)

	router.Handle("POST", "/", postExample)

	// Use AWS Cognito hosted pages for signin
	router.Handle("GET", "/login", redirectToCognitoSignin)
	// Callback for Cognito
	router.Handle("GET", "/callback", cognitoCallback)
	// After login example
	router.Handle("GET", "/userinfo", cognitoProtected, jwtMiddleware)

	// Handle RPCs
	rpcRouter := aegis.NewRPCRouter()
	rpcRouter.Handle("procedure", handleProcedure)

	// Handle S3 objects
	s3Router := aegis.NewS3ObjectRouterForBucket("aegis-incoming")
	// s3Router.Handle("s3:ObjectCreated:Put", "*.png", handleS3Upload)
	// Put() is a shortcut for the above
	s3Router.Put("*.png", handleS3Upload)

	// Handle Cognito Triggers
	cognitoRouter := aegis.NewCognitoRouter()
	cognitoRouter.Handle("PreSignUp_SignUp", handleCognitoPreSignUp)

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
		CognitoRouter:  cognitoRouter,
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
	// The configuration is a function because config can come from a variety of sources
	AegisApp.ConfigureCognitoAppClient(func(ctx context.Context, evt map[string]interface{}) interface{} {
		return &aegis.CognitoAppClientConfig{
			Region:      "us-east-1",
			PoolID:      "xxxxx",
			ClientID:    "xxxxx",
			RedirectURI: "https://xxxxx.amazonaws.com/prod/callback",
		}
	})

	AegisApp.Start()

}

func fallThrough(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	res.StatusCode = 404
	return nil
}

func root(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	aegis.Log.Info("logging to CloudWatch")
	log.Println("normal go logging (also goes to cloudwatch)")

	res.JSON(200, map[string]interface{}{"event": req})
	return nil
}

func postExample(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	form, err := req.GetForm()
	if err != nil {
		res.JSON(500, err.Error())
	} else {
		res.JSON(200, form)
	}
	return nil
}

func somepath(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
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
func fooMiddleware(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) bool {
	log.Println("Foo!")
	return true
}

func barMiddleware(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) bool {
	log.Println("Bar!")
	res.Body = "bar!"
	return true
}

// Redirect to AWS Cognito hosted signin page
func redirectToCognitoSignin(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	log.Println("Redirect to login:", AegisApp.Cognito.HostedLoginURL)
	res.Redirect(301, AegisApp.Cognito.HostedLoginURL)
	// res.JSON(200, map[string]interface{}{"url": CAClient.HostedLoginURL})
	return nil
}

// Handle oauth2 callback, will exchange code for token
func cognitoCallback(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	// Exchange code for token
	tokens, err := AegisApp.Cognito.GetTokens(req.QueryStringParameters["code"], []string{"profile", "openid"})
	if err != nil {
		log.Println("Couldn't get access token", err)
		res.JSONError(500, err)
	} else {
		// verify the token
		_, err := AegisApp.Cognito.ParseAndVerifyJWT(tokens.IDToken)
		if err == nil {
			// Use/send whichever you need for your app
			res.SetHeader("Set-Cookie", "access_token="+tokens.AccessToken+"; Domain=u7aq1oathb.execute-api.us-east-1.amazonaws.com; Secure; HttpOnly")
			// convert to string
			//res.SetHeader("Set-Cookie", "token_expiration="+token.ExpiresIn+"; Domain=u7aq1oathb.execute-api.us-east-1.amazonaws.com; Secure; HttpOnly")
			res.Redirect(301, "https://u7aq1oathb.execute-api.us-east-1.amazonaws.com/prod/userinfo")
		} else {
			res.JSONError(401, errors.New("unauthorized, invalid token"))
		}
	}
	// This is all a one time auth. It verifies the JWT but does nothing else with it.
	// So it won't be held in cookies, etc. Users would need to go back to login all over again for a new code to exchange.
	// See use case 26
	// Better to use this in applications:
	// https://github.com/aws/aws-amplify

	return nil
}

// Example after successful login
func cognitoProtected(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	res.JSON(200, map[string]interface{}{"success": true})
	return nil
}

func jwtMiddleware(ctx context.Context, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) bool {
	jwtCookie, err := req.Cookie("access_token")
	if err != nil {
		log.Println("access_token not found in cookies", err)
		allCookies, _ := req.Cookies()
		log.Println("All cookies:", allCookies)
		return false
	}
	// Check req Host and Referrer for increased protection

	_, err = AegisApp.Cognito.ParseAndVerifyJWT(jwtCookie.Value)
	// parsedToken, err := CAClient.ParseAndVerifyJWT(token.IDToken)
	// if parsedToken.Claims ... some blah blah, or look up some user then if blah blah, then ok.
	if err == nil {
		log.Println("Could not verify JWT", err)
		return true
	}

	// none shall pass
	return false
}

// Example task handler
func handleTask(ctx context.Context, evt *map[string]interface{}) error {
	log.Println("Handling task!", evt)
	return nil
}

// Example task handler catch all
func taskerFallThrough(ctx context.Context, evt *map[string]interface{}) error {
	log.Println("Handling task!", evt)
	return nil
}

// Example RPC handler
func handleProcedure(ctx context.Context, evt map[string]interface{}) (map[string]interface{}, error) {
	log.Println("Handling remote procedure!")
	return evt, nil
}

// Example S3 handler
func handleS3Upload(ctx context.Context, evt *aegis.S3Event) error {
	log.Println("Handling S3 upload!")
	log.Println(evt)
	return nil
}

// Example cognito handler
func handleCognitoPreSignUp(ctx context.Context, evt map[string]interface{}) (map[string]interface{}, error) {
	log.Println("Handling Cognito Pre SignUp!")
	log.Println(evt)
	return evt, nil
}
