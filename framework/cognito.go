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

package framework

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

// CognitoRouter struct provides an interface to handle Cognito trigger events (routers can be for a specific pool or any)
type CognitoRouter struct {
	handlers map[string]CognitoHandler
	PoolID   string
	Tracer   TraceStrategy
}

// CognitoHandler handles routed trigger events, note that these must return a map[string]interface{} response
type CognitoHandler func(context.Context, *HandlerDependencies, map[string]interface{}) (map[string]interface{}, error)

// LambdaHandler handles Cognito trigger events.
func (r *CognitoRouter) LambdaHandler(ctx context.Context, d *HandlerDependencies, evt map[string]interface{}) (map[string]interface{}, error) {
	var err error
	handled := false
	userPoolID := evt["userPoolId"].(string)
	triggerSource := evt["triggerSource"].(string)
	userName := evt["userName"].(string)
	// The trigger typically comes with a default response.
	// Maybe that can be used if an empty result is returned from handler??
	// What if an empty map is intended?
	// defaultResponse := evt["response"].(map[string]interface{})
	var response map[string]interface{}

	// If an incoming event can be matched to this router, but the router has no registered handlers
	// or if one hasn't been added to aegis.Handlers{}.
	if r == nil {
		return response, errors.New("no handlers registered for CognitoRouter")
	}

	if r.PoolID == "" || r.PoolID == userPoolID {
		if handler, ok := r.handlers[triggerSource]; ok {
			handled = true
			r.Tracer.Record("annotation",
				map[string]interface{}{
					"CognitoUserPoolID":    userPoolID,
					"CognitoTriggerSource": triggerSource,
					"UserName":             userName,
				},
			)

			err = r.Tracer.Capture(ctx, "CognitoHandler", func(ctx1 context.Context) error {
				d.Tracer = &r.Tracer
				response, err = handler(ctx1, d, evt)
				return err
			})
		}
	}

	// Otherwise, use the catch all (router "fallthrough" equivalent) handler.
	// The application can inspect the map and make a decision on what to do, if anything.
	// This is optional.
	if !handled {
		log.Println("using default fall through handler")
		// It's possible that the CognitoRouter wasn't created with NewCognitoRouter, so check for this still.
		if handler, ok := r.handlers["_"]; ok {
			// Capture the handler (in XRay by default) automatically
			r.Tracer.Record("annotation",
				map[string]interface{}{
					"CognitoUserPoolID":    userPoolID,
					"CognitoTriggerSource": triggerSource,
					"UserName":             userName,
					"FallthroughHandler":   true,
				},
			)

			err = r.Tracer.Capture(ctx, "CognitoHandler", func(ctx1 context.Context) error {
				d.Tracer = &r.Tracer
				response, err = handler(ctx, d, evt)
				return err
			})
		}
	}

	return response, err
}

// Listen will start an S3 even listener that handles incoming object based events (put, delete, etc.)
func (r *CognitoRouter) Listen() {
	lambda.Start(r.LambdaHandler)
}

// NewCognitoRouter simply returns a new CognitoRouter struct and behaves a bit like Router, it even takes an optional rootHandler or "fall through" catch all
func NewCognitoRouter(rootHandler ...CognitoHandler) *CognitoRouter {
	// The catch all is optional, if not provided, an empty handler is still called and it returns nothing.
	handler := func(context.Context, *HandlerDependencies, map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{}, nil
	}
	if len(rootHandler) > 0 {
		handler = rootHandler[0]
	}
	return &CognitoRouter{
		handlers: map[string]CognitoHandler{
			"_": handler,
		},
	}
}

// NewCognitoRouterForPool is the same as NewCognitoRouter except it's for a specific pool ID (you could also set the PoolID field after using the other function)
func NewCognitoRouterForPool(poolID string, rootHandler ...CognitoHandler) *CognitoRouter {
	var r *CognitoRouter
	if len(rootHandler) > 0 {
		r = NewCognitoRouter(rootHandler[0])
	} else {
		r = NewCognitoRouter()
	}
	// Just convenience
	r.PoolID = poolID
	return r
}

// Handle will register a handler for a given Cognito trigger source
// List of triggerSources here: https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-pools-lambda-trigger-syntax-shared.html
func (r *CognitoRouter) Handle(triggerSource string, handler func(context.Context, *HandlerDependencies, map[string]interface{}) (map[string]interface{}, error)) {
	if r.handlers == nil {
		r.handlers = make(map[string]CognitoHandler)
	}
	r.handlers[triggerSource] = handler
}

// PreSignUp is the same as Handle only the triggerSource is already implied. It handles any PreSignUp_SignUp event.
func (r *CognitoRouter) PreSignUp(handler func(context.Context, *HandlerDependencies, map[string]interface{}) (map[string]interface{}, error)) {
	r.Handle("PreSignUp_SignUp", handler)
}

// Too many?
// CustomMessage_ResendCode
// func (r *CognitoRouter) ResendCodeMessage(handler func(context.Context, *map[string]interface{}) error) {
// 	r.Handle("CustomMessage_ResendCode", handler)
// }

// PreAuthentication_Authentication
