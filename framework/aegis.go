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
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	lambdaSDK "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	prettyjson "github.com/hokaccha/go-prettyjson"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
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

	// SimpleEmailEvent alias for SES Email events (recipient rules)
	SimpleEmailEvent events.SimpleEmailEvent
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
	Custom          map[string]interface{}
	Services
	Filters struct {
		Handler struct {
			BeforeServices []func(*context.Context, map[string]interface{})
			Before         []func(*context.Context, map[string]interface{})
			After          []func(*context.Context, interface{}, error) (interface{}, error)
		}
	}
}

// Services defines core framework services such as auth
type Services struct {
	Cognito        *CognitoAppClient
	Variables      map[string]string
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

// setAegisVariables will set Aegis variables (on `a.Services.Variables`) to be used by Services and user handler code.
func (a *Aegis) setAegisVariables(ctx context.Context, evt map[string]interface{}) {
	// Aegis variables follow a convention for being set.
	// First the Lambda environment variables are used.
	// Then the API Gateway stage variables.
	// So stage variables can override keys from Lambda environment variables.
	// However, both are still very available if the user wants to look them up.
	// A a developer/user, it'll be important to understand where to set variables depending on the event type/handler as well
	// as the architectural strategy (one function per "environment" with multiple API stages, or multiple functions/gateways?).

	if a.Services.Variables == nil {
		a.Services.Variables = make(map[string]string)
	}

	// Lambda environment variables first
	lc, _ := lambdacontext.FromContext(ctx)
	if lc != nil && len(lc.ClientContext.Env) > 0 {
		for k, v := range lc.ClientContext.Env {
			a.Services.Variables[k] = v
		}
	}

	// API Gateway stage variables can override values set by Lambda environment variables
	if getType(evt) == "APIGatewayProxyRequest" {
		var e APIGatewayProxyRequest
		// The event contains no time/date, should decode just fine
		err := mapstructure.Decode(evt, &e)
		if err == nil && len(e.StageVariables) > 0 {
			for k, v := range e.StageVariables {
				a.Services.Variables[k] = v
				// Try to base64 decode it, because API Gateway stage variables may be encoded because
				// they do not support certain special characters, which is a big problem for sensitive
				// credentials that often include special characters.
				sDec, err := base64.StdEncoding.DecodeString(v)
				if err == nil {
					a.Services.Variables[k] = string(sDec)
				}
			}
		}
	}
}

// aegisHandler configures services and determines how to handle the Lambda event
func (a *Aegis) aegisHandler(ctx context.Context, evt map[string]interface{}) (interface{}, error) {
	if a.TraceContext == nil {
		a.TraceContext = ctx
	}

	// Set "Aegis Variables" for use by both service configurations and handlers.
	a.setAegisVariables(ctx, evt)

	// Filters to run before anything is handled, even before services are configured.
	if a.Filters.Handler.BeforeServices != nil {
		for _, filter := range a.Filters.Handler.BeforeServices {
			filter(&ctx, evt)
		}
	}

	// If a "cognito" configuration function was provided and Cognito has not been configured already
	if sCfg, ok := a.Services.configurations["cognito"]; ok && a.Services.Cognito == nil {
		cognitoCfg := sCfg(ctx, evt).(*CognitoAppClientConfig)
		cognitoCfg.TraceContext = a.TraceContext
		cognitoCfg.AWSClientTracer = a.AWSClientTracer

		a.Tracer.Record("annotation",
			map[string]interface{}{
				"CognitoRegion":         cognitoCfg.Region,
				"CognitoUserPoolID":     cognitoCfg.PoolID,
				"CognitoAppClientID":    cognitoCfg.ClientID,
				"CognitoAppRedirectURI": cognitoCfg.RedirectURI,
			},
		)

		err := a.Tracer.Capture(ctx, "NewCognitoAppClient", func(ctx1 context.Context) error {
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
			filter(&ctx, evt)
		}
	}

	// Dependencies to be injected into each event handler
	// Note that custom "user" dependencies can be passed from Aegis interface down to each handler
	d := HandlerDependencies{
		Services: &a.Services,
		Log:      a.Log,
		Tracer:   &a.Tracer,
		Custom:   a.Custom,
	}

	// This could be called directly of course, it would skip all of the service set up (if there were any configured)
	res, err := a.Handlers.eventHandler(ctx, &d, evt)

	// Filters to run after handling the event. Instead of getting a map[string]interface{} with the event,
	// this filter gets the interface{} and error that would normally be returned. It presents an opportunity
	// to change the entire return values.
	if a.Filters.Handler.After != nil {
		for _, filter := range a.Filters.Handler.After {
			res, err = filter(&ctx, res, err)
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

// standAloneHandler implements an http.Handler and bring Aegis along for handling events
type standAloneHandler struct {
	aegis *Aegis
	cfg   *StandAloneCfg
}

// StandAloneCfg allows the local HTTP server to be configured, passed to StartServer()
type StandAloneCfg struct {
	Port              string
	AllowHeaders      []string
	AllowMethods      []string
	AllowOrigin       string
	AdditionalHeaders map[string]string
	CLI               StandAloneCLICfg
}

// StandAloneCLICfg has some options for the CLI when running "stand alone" functions
type StandAloneCLICfg struct {
	PrettyPrint   bool
	HideLogOutput bool
}

// ServeHTTP will start an HTTP listener locally, to handle incoming requests (mainly intended for local development)
func (h standAloneHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// --> Build the request by converting HTTP request to Lambda Event
	ctx, req := h.requestToProxyRequest(r)

	// We need the event to come from JSON for the field names (ie. "httpMethod" instead of struct field HTTPMethod)
	// So just marshal to JSON
	evtJSON, _ := json.Marshal(req)
	// log.Println(string(evtJSON))
	// and unmarshal the event as map[string]interface{}
	// It's a little repetitive, aegisHandler() will convert to struct again.
	// The other option is to have requestToProxyRequest() return a map[string]interface{} instead.
	// Unsure how much it affects performance and if performance is even a concern here since the intent
	// is to provide a way to test an API locally as one develops it.
	var evt map[string]interface{}
	if err := json.Unmarshal(evtJSON, &evt); err != nil {
		panic(err)
	}

	// NOTE: Tracer could be configured/set to a custom Tracer interface that doesn't use xray at all.
	tracedCtx, _ := h.aegis.Tracer.BeginSegment(ctx, "Aegis")

	// -<>- Normal handling of Lambda Event with Aegis Router.
	resp, err := h.aegis.aegisHandler(tracedCtx, evt)

	if apiResp, ok := resp.(APIGatewayProxyResponse); ok {
		// <-- Send the response (convert standard returned map[string]interface{} to APIGatewayProxyResponse)
		h.proxyResponseToHTTPResponse(&apiResp, err, w)
	} else {
		w.WriteHeader(500)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		} else {
			fmt.Fprintf(w, "Invalid request.")
		}
	}
}

// requestToProxyRequest will take an HTTP Request and transform it into a faux AWS Lambda events.APIGatewayProxyRequest.
// The APIGatewayProxyRequestContext will be missing some data, such as AccountID. So any route handlers that depend
// on that information may not work locally as expect. However, this will allow us to run a local web server for the API.
// This is mainly useful for local development and testing.
func (h standAloneHandler) requestToProxyRequest(r *http.Request) (context.Context, *APIGatewayProxyRequest) {
	ctx := context.Background()
	req := APIGatewayProxyRequest{
		Path:       r.URL.Path,
		HTTPMethod: r.Method,
		RequestContext: events.APIGatewayProxyRequestContext{
			HTTPMethod: r.Method,
		},
	}

	// transfer the headers over to the event
	req.Headers = map[string]string{}
	for k, v := range r.Header {
		req.Headers[k] = strings.Join(v, "; ")
	}

	// Querystring params
	params := r.URL.Query()
	paramsMap := map[string]string{}
	for k, _ := range params {
		paramsMap[k] = params.Get(k)
	}
	req.QueryStringParameters = paramsMap

	// Path params (just the proxy+ path ... but it does not have the preceding slash)
	req.PathParameters = map[string]string{
		"proxy": r.URL.Path[1:len(r.URL.Path)],
	}
	req.Resource = "/{proxy+}"
	req.RequestContext.ResourcePath = "/{proxy+}"

	// Identity info: user agent, IP, etc.
	req.RequestContext.Identity.UserAgent = r.Header.Get("User-Agent")
	req.RequestContext.Identity.SourceIP = r.Header.Get("X-Forwarded-For")
	if req.RequestContext.Identity.SourceIP == "" {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			req.RequestContext.Identity.SourceIP = net.ParseIP(ip).String()
		}
	}

	// Stage will be "local" for now? I'm not sure what makes sense here. Local gateway. Local. Debug. ¯\_(ツ)_/¯
	req.RequestContext.Stage = "local"
	// TODO: Stage variables would need to be pulled from the aegis.yaml ...
	// so now the config file has to be next to the app... otherwise some defaults will be set like "local"
	// and no stage variables i suppose.
	// evt.StageVariables =

	// The request id will simply be a timestamp to help keep it unique, but also allowing it to be easily sorted
	req.RequestContext.RequestID = strconv.FormatInt(time.Now().UnixNano(), 10)

	// pass along the body
	bodyData, err := ioutil.ReadAll(r.Body)
	if err == nil {
		req.Body = string(bodyData)
	}

	return ctx, &req
}

// proxyResponseToHTTPResponse will take the typical Lambda Proxy response and transform it into an HTTP response.
// AWS does this for us automatically, but when running a local HTTP server, we'll need to do it.
func (h standAloneHandler) proxyResponseToHTTPResponse(res *APIGatewayProxyResponse, err error, w http.ResponseWriter) {
	// transfer the headers into the HTTP Response
	if res.Headers != nil {
		for k, v := range res.Headers {
			w.Header().Set(k, v)
		}
	}

	// TODO: Actually allow this to be configured.
	// CORS. Allow everything since we are assumed to be running locally.
	allowedHeaders := []string{
		"Accept",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"Authorization",
		"X-CSRF-Token",
		"X-Auth-Token",
	}
	allowedMethods := []string{
		"POST",
		"GET",
		"OPTIONS",
		"PUT",
		"PATCH",
		"DELETE",
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))

	// err is any error returned from a handler handling an event.
	// If it isn't nil, then we can't even attempt to return the response, it might not be set.
	// We should return the error, but also set the header to 500.
	// TODO: Question is, should this be plain text or JSON? Hard to say what is consuming it.
	// Maybe also configure that.
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, err.Error())
	} else {
		// Return a successful response

		// The handler and middleware should have set everything on res
		w.WriteHeader(res.StatusCode)

		// If this is true, then API Gateway will decode the base64 string to bytes. Mimic that behavior here.
		if res.IsBase64Encoded {
			decodedBody, err := base64.StdEncoding.DecodeString(res.Body)
			if err == nil {
				w.Header().Set("Content-Length", strconv.Itoa(len(decodedBody)))
				fmt.Fprintf(w, "%s", decodedBody)
			} else {
				res.Body = err.Error()
				fmt.Fprintf(w, res.Body)
			}
		} else {
			// if not base64, write res.Body
			fmt.Fprintf(w, res.Body)
		}
	}
}

// StartServer will tell all handlers to listen for events, but will do so through a local HTTP server
// This allows for easy local dev or even running functions outside of AWS Lambda
func (a *Aegis) StartServer(cfg ...StandAloneCfg) {
	// Default config if one hasn't been provided
	localCfg := StandAloneCfg{
		Port: "9999",
		AllowHeaders: []string{
			"Accept",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
			"X-CSRF-Token",
			"X-Auth-Token",
		},
		AllowMethods: []string{
			"POST",
			"GET",
			"OPTIONS",
			"PUT",
			"PATCH",
			"DELETE",
		},
	}
	// Overrides default
	if len(cfg) > 0 {
		localCfg = cfg[0]
		// Ensure there's a port defined though, just in cases
		if localCfg.Port == "" {
			localCfg.Port = "9999"
		}
	}

	httpPort := ":" + localCfg.Port
	log.Printf("Starting local gateway: http://localhost%v \n", httpPort)

	err := http.ListenAndServe(httpPort, &standAloneHandler{aegis: a, cfg: &localCfg})
	if err != nil {
		log.Fatal(err)
	}
}

// StartSingle will tell all handlers to listen for events, but will receive its event via command line flag
func (a *Aegis) StartSingle(cfg ...StandAloneCfg) {
	var eventPtr = flag.String("event", "", "path/to/event.json")
	var prettyPtr = flag.Bool("pretty", false, "pretty print the response")
	var nologPtr = flag.Bool("nolog", false, "hide log output, only show result")
	flag.Parse()

	singleCfg := StandAloneCfg{
		CLI: StandAloneCLICfg{
			PrettyPrint: false,
		},
	}
	if len(cfg) > 0 {
		singleCfg = cfg[0]
	}

	// Obey optional hide log flag over config, if set
	if nologPtr != nil {
		nolog := *nologPtr
		singleCfg.CLI.HideLogOutput = nolog
	}

	// Obey optional flag over config, if set
	if prettyPtr != nil {
		prettyPrint := *prettyPtr
		singleCfg.CLI.PrettyPrint = prettyPrint
	}

	// Hide log output if set, this can be helpful if the function has a lot of noisy logging
	// and only the response/error is desired
	if singleCfg.CLI.HideLogOutput {
		log.SetOutput(ioutil.Discard)
		a.Log.Out = ioutil.Discard
	}

	// If the event file path was passed
	if eventPtr != nil {
		eventPath := *eventPtr
		// fmt.Println("event file path:", eventPath)
		// TODO: Maybe if its a directory, loop all files in it?
		// Use that instead of StartSuite() ?
		// Though such a StartSuite() command might also have more handy info
		// with a summary and not exit on failure.

		raw, err := ioutil.ReadFile(eventPath)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		var evt map[string]interface{}
		json.Unmarshal(raw, &evt)

		// NOTE: Tracer could be configured/set to a custom Tracer interface that doesn't use xray at all.
		ctx := context.Background()
		tracedCtx, _ := a.Tracer.BeginSegment(ctx, "Aegis")

		// -<>- Normal handling of Lambda Event with Aegis Router.
		resp, err := a.aegisHandler(tracedCtx, evt)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		} else {
			// Pretty print is optional
			if singleCfg.CLI.PrettyPrint {
				// APIGatewayProxyResponse types will have escaped Body value when pretty printed.
				// Handle that and pretty print the body as well.
				if apiResp, ok := resp.(APIGatewayProxyResponse); ok {
					// If it can be formatted
					prettyBody, err := prettyjson.Format([]byte(apiResp.Body))

					if err == nil {
						apiResp.Body = ""
						s, _ := prettyjson.Marshal(apiResp)
						fmt.Println("\nAPIGatewayProxyResponse:")
						fmt.Println(string(s))

						fmt.Println("\nBody (JSON):")
						fmt.Println(string(prettyBody))
					} else {
						s, _ := prettyjson.Marshal(resp)
						fmt.Println(string(s))
					}
				} else {
					s, _ := prettyjson.Marshal(resp)
					fmt.Println(string(s))
				}
			} else {
				fmt.Print(resp)
			}

			// Exit ok.
			os.Exit(0)
		}
	}
}

// possibly TODO: StartSuite() ... a way of handling a set of events from a directory like maybe `_integration`
// it will loop over the files with JSON in there and run them one by one returning output to the CLI.
// this could be a useful feature for CI/CD. But keep in mind that mock services are likely needed.
// it's very difficult to just trigger a fake S3 Put event for example, because the handler likely wants
// to go retrieve that file from S3 to do something and so the event itself could point to a file that
// doesn't exist. So each type of event will be a special consideration and likely need mock data or mocked
// services. However, it's a nice feature to have if one takes the time to set up integration tests.
// A function like this would provide a very easy way for a CI/CD tool to build and run the binary
// and then only deploy if it passes these tests. Which could be in addition to normal go test.
