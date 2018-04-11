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

// Parts borrowed from https://github.com/acmacalister/helm

package framework

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	get     = "GET"
	head    = "HEAD"
	post    = "POST"
	put     = "PUT"
	patch   = "PATCH"
	delete  = "DELETE"
	options = "OPTIONS"
)

// RouteHandler is similar to "net/http" Handler, except there is no response writer.
// Instead the *APIGatewayProxyResponse is manipulated and returned directly.
type RouteHandler func(context.Context, *HandlerDependencies, *APIGatewayProxyRequest, *APIGatewayProxyResponse, url.Values) error

// Middleware is just like the RouteHandler type, but has a boolean return. True
// means to keep processing the rest of the middleware chain, false means end.
type Middleware func(context.Context, *HandlerDependencies, *APIGatewayProxyRequest, *APIGatewayProxyResponse, url.Values) bool

// Router name says it all.
type Router struct {
	tree           *node
	rootHandler    RouteHandler
	middleware     []Middleware
	l              *log.Logger
	LoggingEnabled bool
	URIVersion     string
	GatewayPort    string
	Tracer         TraceStrategy
}

var (
	// ErrNameNotProvided is thrown when a name is not provided
	ErrNameNotProvided = errors.New("no name was provided in the HTTP body")
)

// NewRouter creates a new router. Take the root/fall through route
// like how the default mux works. Only difference is in this case,
// you have to specific one.
func NewRouter(rootHandler RouteHandler) *Router {
	node := node{component: "/", isNamedParam: false, methods: make(map[string]*route)}
	return &Router{tree: &node, rootHandler: rootHandler, URIVersion: ""}
}

// Handle takes an http handler, method and pattern for a route.
func (r *Router) Handle(method, path string, handler RouteHandler, middleware ...Middleware) {
	if path[0] != '/' {
		panic("Path has to start with a /.")
	}
	r.tree.addNode(method, r.URIVersion+path, handler, middleware...)
}

// GET same as Handle only the method is already implied.
func (r *Router) GET(path string, handler RouteHandler, middleware ...Middleware) {
	r.Handle(get, path, handler, middleware...)
}

// HEAD same as Handle only the method is already implied.
func (r *Router) HEAD(path string, handler RouteHandler, middleware ...Middleware) {
	r.Handle(head, path, handler, middleware...)
}

// OPTIONS same as Handle only the method is already implied.
func (r *Router) OPTIONS(path string, handler RouteHandler, middleware ...Middleware) {
	r.Handle(options, path, handler, middleware...)
}

// POST same as Handle only the method is already implied.
func (r *Router) POST(path string, handler RouteHandler, middleware ...Middleware) {
	r.Handle(post, path, handler, middleware...)
}

// PUT same as Handle only the method is already implied.
func (r *Router) PUT(path string, handler RouteHandler, middleware ...Middleware) {
	r.Handle(put, path, handler, middleware...)
}

// PATCH same as Handle only the method is already implied.
func (r *Router) PATCH(path string, handler RouteHandler, middleware ...Middleware) {
	r.Handle(patch, path, handler, middleware...)
}

// DELETE same as Handle only the method is already implied.
func (r *Router) DELETE(path string, handler RouteHandler, middleware ...Middleware) {
	r.Handle(delete, path, handler, middleware...)
}

// runMiddleware loops over the slice of middleware and call to each of the middleware handlers.
func runMiddleware(ctx context.Context, d *HandlerDependencies, req *APIGatewayProxyRequest, res *APIGatewayProxyResponse, params url.Values, middleware ...Middleware) bool {
	for _, m := range middleware {
		if !m(ctx, d, req, res, params) {
			return false // the middleware returned false, so end processing the chain.
		}
	}
	return true
}

// LambdaHandler is a native AWS Lambda Go handler function (no more shim).
func (r *Router) LambdaHandler(ctx context.Context, d *HandlerDependencies, req APIGatewayProxyRequest) (APIGatewayProxyResponse, error) {
	// url.Values are typically used for qureystring parameters.
	// However, this router uses them for path params.
	// Querystring parameters can be picked up from the *Event though.
	params := url.Values{}

	var res APIGatewayProxyResponse
	var err error

	// use the Path and HTTPMethod from the event to figure out the route
	node, _ := r.tree.traverse(strings.Split(req.Path, "/")[1:], params)
	if handler := node.methods[req.HTTPMethod]; handler != nil {
		// Middleware must return true in order to continue.
		// If it returns false, it will catch and halt everything.
		if !runMiddleware(ctx, d, &req, &res, params, handler.middleware...) {
			// Return the response in its current stage if middleware returns false.
			// It is up to the middleware itself to set the response returned.
			// Maybe some authentication failed? So maybe the middleware wants to return a message about that.
			// But since it failed, it will not proceed any farther with the next middleware or route handler.
			return res, nil
		}
		// Trac/capture the handler (in XRay by default) automatically
		r.Tracer.Annotations = map[string]interface{}{
			"RequestPath": req.Path,
			"Method":      req.HTTPMethod,
		}
		err = r.Tracer.Capture(ctx, "RouteHandler", func(ctx1 context.Context) error {
			r.Tracer.AddAnnotations(ctx1)
			r.Tracer.AddMetadata(ctx1)

			// Set the injected tracer to this router Tracer (was Aegis interface's tracer).
			// This is important. It allows annotations to be added by handlers to be traced automatically.
			// This means the end user does not need to set up their own tracer. They can hook into the current trace.
			d.Tracer = &r.Tracer
			// I believe ctx1 is actually the same as ctx in this case. Capture() makes no copy of context.
			// Context is immutable. So... To not be confusing, we'll use ctx1.
			return handler.handler(ctx1, d, &req, &res, params)
		})

		// TODO: look at environment variable to see if XRay was disabled (env var on lambda or when running local server)
		// Then just call handler and not the xray part above.
		// handler.handler(ctx, &req, &res, params)
	} else {
		r.rootHandler(ctx, d, &req, &res, params)
	}

	// Returning an error from this handler is how AWS Lambda works, but when dealing with API Gateway, it doesn't make for
	// a great response. A 502 Bad Gateway is returned and a JSON message that isn't helpful or adjustable. So we can simply
	// handle all errors this way and never return an error from the Lambda handler itself when using the Router.
	if err != nil {
		// At this point, the handler may have set a content type header.
		// This function will look at that when determining how to display the error in the response body.
		// TODO: Allow the error responses to use a template defined via configuration.
		// This would allow an application to change the body response format to suit its needs and still use the errors for XRay.
		// At least for now, the response will be of the intended type and should contain something somewhat useful.
		res.Error(500, err)
	}
	return res, nil
}

// Listen will start the internal router and listen for Lambda events to forward to registered routes.
func (r *Router) Listen() {
	lambda.Start(r.LambdaHandler)
}

// gatewayHandler is a Router that implements an http.Handler interface
type gatewayHandler Router

// ServerHTTP will build an Event using an HTTP Request, run the appropriate handler from the Lambda code and then return an HTTP Response.
// Normally AWS API Gateway handles this for us, but locally, we'll need to do these transforms.
func (h gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// --> Build the request by converting HTTP request to Lambda Event
	ctx, req := h.requestToProxyRequest(r)
	// TODO: Move all this to Aegis interface so we can add dependencies
	d := &HandlerDependencies{}

	// -<>- Normal handling of Lambda Event with Aegis Router.
	// It looks just like the closure given to RunStream with Listen(), except that we aren't going
	// to deal with stdio. Instead, we're taking the results and passing back through an HTTP Response.
	// TODO: refactor this to make a new handler that also works with a router. There's a tiny bit of repetition with this and what's in Listen()
	params := url.Values{}
	var res APIGatewayProxyResponse
	// use the Path and HTTPMethod from the event to figure out the route
	node, _ := h.tree.traverse(strings.Split(req.Path, "/")[1:], params)
	if handler := node.methods[req.HTTPMethod]; handler != nil {
		// Middleware must return true in order to continue.
		// If it returns false, it will catch and halt everything.
		if !runMiddleware(ctx, d, req, &res, params, handler.middleware...) {
			// <-- Send the response
			h.proxyResponseToHTTPResponse(&res, w)
			return
		}
		handler.handler(ctx, d, req, &res, params)
	} else {
		h.rootHandler(ctx, d, req, &res, params)
	}

	// <-- Send the response
	h.proxyResponseToHTTPResponse(&res, w)
}

// requestToProxyRequest will take an HTTP Request and transform it into a faux AWS Lambda events.APIGatewayProxyRequest.
// The APIGatewayProxyRequestContext will be missing some data, such as AccountID. So any route handlers that depend
// on that information may not work locally as expect. However, this will allow us to run a local web server for the API.
// This is mainly useful for local development and testing.
func (h gatewayHandler) requestToProxyRequest(r *http.Request) (context.Context, *APIGatewayProxyRequest) {
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
func (h gatewayHandler) proxyResponseToHTTPResponse(res *APIGatewayProxyResponse, w http.ResponseWriter) {
	// transfer the headers into the HTTP Response
	for k, v := range res.Headers {
		w.Header().Set(k, v)
	}

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

// Gateway will start a local web server to listen for events. Useful for testing locally.
func (r *Router) Gateway() {
	if r.GatewayPort == "" {
		r.GatewayPort = ":9999"
	}
	log.Printf("Starting local gateway: http://localhost%v \n", r.GatewayPort)
	err := http.ListenAndServe(r.GatewayPort, gatewayHandler(*r))
	if err != nil {
		log.Fatal(err)
	}
}
