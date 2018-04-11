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
	"context"
	"encoding/json"
	"log"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	lambdaSDK "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
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

	// S3Event alias
	S3Event events.S3Event

	// CognitoEvent alias (NOT a Cognito Trigger event, this is for sync)
	CognitoEvent events.CognitoEvent

	// CloudWatchEvent alias for CloudWatchEvent events
	CloudWatchEvent events.CloudWatchEvent
)

// Log uses Logrus for logging and will hook to CloudWatch...But could also be used to hook to other centralized logging services.
var Log = logrus.New()

// Aegis is the framework's super interface, it holds various configurations and services
// While it is possible to use many of the framework's interfaces and routers/handlers individually, it's often
// more convenient to use this main interface to avoid unnecessary SDK calls, etc.
type Aegis struct {
	Handlers
	Log             *logrus.Logger
	AWSClientTracer func(c *client.Client)
	Tracer          TraceStrategy
	TraceContext    context.Context
	Services
	Filters struct {
		Handler struct {
			BeforeServices []func(*context.Context, *map[string]interface{})
			Before         []func(*context.Context, *map[string]interface{})
			After          []func(*context.Context, *interface{})
		}
	}
}

// Services defines core framework services such as auth
type Services struct {
	Cognito        *CognitoAppClient
	configurations map[string]func(context.Context, map[string]interface{}) interface{}
}

// New will return a new Aegis interface with handlers (many, but not all, handlers are routers with many handlers)
func New(handlers Handlers) *Aegis {
	return &Aegis{
		Handlers: handlers,
		Log:      logrus.New(),
		Services: Services{
			configurations: make(map[string]func(context.Context, map[string]interface{}) interface{}),
		},
		AWSClientTracer: xray.AWS,
	}
}

// ConfigureLogger will set a custom logrus Logger, overriding the default which just goes to stdout (and to CloudWatch).
// Logrus has been chosen for it's pluggability and extensive list of hooks. Send to Bugsnag, Fluentd, InfluxDB, Slack and more.
// See: https://github.com/sirupsen/logrus
func (a *Aegis) ConfigureLogger(logger *logrus.Logger) {
	// Available as both framework.Log and on Aegis interface
	Log = logger
	a.Log = Log
}

// Start will tell all handlers to listen for events, it's designed to be similar to lambda.Start()
func (a *Aegis) Start() {
	lambda.Start(a.aegisHandler)
}

// ConfigureService will configure an AegisService
func (a *Aegis) ConfigureService(name string, cfg func(context.Context, map[string]interface{}) interface{}) {
	a.Services.configurations[name] = cfg
}

// aegisHandler configures services and determines how to handle the Lambda event
func (a *Aegis) aegisHandler(ctx context.Context, evt map[string]interface{}) (interface{}, error) {
	if a.TraceContext == nil {
		a.TraceContext = ctx
	}

	// Filters to run before anything is handled, even before services are configured.
	if a.Filters.Handler.BeforeServices != nil {
		for _, filter := range a.Filters.Handler.BeforeServices {
			filter(&ctx, &evt)
		}
	}

	// If a "cognito" configuration function was provided and Cognito has not been configured already
	if sCfg, ok := a.Services.configurations["cognito"]; ok && a.Services.Cognito == nil {
		cognitoCfg := sCfg(ctx, evt).(*CognitoAppClientConfig)
		cognitoCfg.TraceContext = a.TraceContext
		cognitoCfg.AWSClientTracer = a.AWSClientTracer

		a.Tracer.Annotations = map[string]interface{}{
			"CognitoRegion":         cognitoCfg.Region,
			"CognitoUserPoolID":     cognitoCfg.PoolID,
			"CognitoAppClientID":    cognitoCfg.ClientID,
			"CognitoAppRedirectURI": cognitoCfg.RedirectURI,
		}
		err := a.Tracer.Capture(ctx, "NewCognitoAppClient", func(ctx1 context.Context) error {
			a.Tracer.AddAnnotations(ctx1)
			a.Tracer.AddMetadata(ctx1)

			// Configure Cognito App Client and set on Aegis struct
			svc, err := NewCognitoAppClient(cognitoCfg)
			a.Services.Cognito = svc
			return err
		})

		if err != nil {
			log.Println("Cognito app client could not be configured")
			log.Println(err)
		}
	}

	// Filters to run before handling the event (but after services have been configured).
	if a.Filters.Handler.Before != nil {
		for _, filter := range a.Filters.Handler.Before {
			filter(&ctx, &evt)
		}
	}

	// Dependencies to be injected into each event handler
	d := HandlerDependencies{
		Services: &a.Services,
		Log:      a.Log,
		Tracer:   &a.Tracer,
	}

	// This could be called directly of course, it would skip all of the service set up (if there were any configured)
	res, err := a.Handlers.eventHandler(ctx, &d, evt)

	// Filters to run after handling the event. Instead of getting a map[string]interface{} with the event,
	// this filter gets an interface{} that is the response.
	if a.Filters.Handler.After != nil {
		for _, filter := range a.Filters.Handler.After {
			filter(&ctx, &res)
		}
	}

	return res, err
}

// RPC makes an Aegis remote procedure call (invokes another Lambda) with tracing support
func (a *Aegis) RPC(functionName string, message map[string]interface{}) (map[string]interface{}, error) {
	sess, err := session.NewSession()
	if err != nil {
		log.Println("could not make remote procedure call, session could not be created")
		return nil, err
	}

	// Payload will need JSON bytes
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("could not marshal remote procedure call message")
		return nil, err
	}

	// region? cross account?
	svc := lambdaSDK.New(sess)
	a.AWSClientTracer(svc.Client)

	// TODO: Look into this more. So many interesting options here. InvocationType and LogType could be interesting outside of defaults
	output, err := svc.InvokeWithContext(a.TraceContext, &lambdaSDK.InvokeInput{
		// ClientContext // TODO: think about this...
		FunctionName: aws.String(functionName),
		// JSON bytes, sadly it does not pass just any old byte array. It's going to come in as a map to the handler.
		// That's a map[string]interface{} from JSON. I saw byte array at first and got excited.
		Payload: jsonBytes,
		// Qualifier ... this is an interesting one. We use latest by default...But we also want to work in a circuit breaker
		// here, so we'll need to set the qualifier at some point.
	})

	// Unmarshal response.
	var resp map[string]interface{}
	if err == nil {
		err = json.Unmarshal(output.Payload, &resp)
	}

	return resp, err
}
