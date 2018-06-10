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

// Parts borrowed from https://github.com/acmacalister/helm

package framework

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/justinas/alice"
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
	stdMiddleware  []func(h http.Handler) http.Handler
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

// Use will set middleware on the Router that gets used by all handled routes.
func (r *Router) Use(middleware ...Middleware) {
	r.middleware = append(r.middleware, middleware...)
}

// UseStandard will set stdMiddleware on the Router that gets used by all handled routes.
// This is standard, idiomatic, Go http.Handler middleware.
func (r *Router) UseStandard(middleware ...func(h http.Handler) http.Handler) {
	r.stdMiddleware = append(r.stdMiddleware, middleware...)
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

// runStandardMiddleware will proxy events to standard http so that standard middleware can be used, then return back.
func runStandardMiddleware(ctx context.Context, req *APIGatewayProxyRequest, middleware ...func(http.Handler) http.Handler) (APIGatewayProxyResponse, error) {
	var res APIGatewayProxyResponse
	var err error

	if middleware != nil && len(middleware) > 0 {
		var constructors []alice.Constructor
		for _, m := range middleware {
			constructors = append(constructors, m)
		}

		emptyHandler := func(w http.ResponseWriter, req *http.Request) {
		}
		adapter := NewHandlerAdapter(emptyHandler)

		// Use a small helper like Alice here
		chained := alice.New(constructors...).Then(adapter.HandlerFunc)
		// Apollo is a fork of Alice and might also be nice if we want context to pass through.
		// Although the Proxy() function will take context and will apply it to the http Request.
		// So it's available to middleware, it just isn't part of the function signature.
		// chained := apollo.New(constructors...).With(ctx).Then(adapter.HandlerFunc)
		adapter.Handler = chained

		// proxyResp, err := adapter.Proxy(events.APIGatewayProxyRequest(*req)) // <-- empty handler by itself
		proxyResp, err := adapter.Proxy(ctx, events.APIGatewayProxyRequest(*req)) // <-- alice chained middleware plus empty handler

		if err == nil {
			res = APIGatewayProxyResponse(proxyResp)
		} else {
			log.Println("Error getting response from adapter.Proxy()", err)
		}
	}

	return res, err
}

// LambdaHandler is a native AWS Lambda Go handler function (no more shim).
func (r *Router) LambdaHandler(ctx context.Context, d *HandlerDependencies, req APIGatewayProxyRequest) (APIGatewayProxyResponse, error) {
	// If an incoming event can be matched to this router, but the router has no registered handlers
	// or if one hasn't been added to aegis.Handlers{}.
	if r == nil {
		return APIGatewayProxyResponse{}, errors.New("no handlers registered for Router")
	}

	// url.Values are typically used for qureystring parameters.
	// However, this router uses them for path params.
	// Querystring parameters can be picked up from the *Event though.
	params := url.Values{}

	// These used to be just declared here. But now runStandardMiddleware() will declare them
	// since it runs first and has to run first. If not using any standard middleware, it will
	// effectively be no different than what it used to be.
	// var res APIGatewayProxyResponse
	// var err error

	// First run the standard http middleware added with UseStandard().
	// Note: Standard http middleawre does not get nearly as many args as Aegis middleware.
	// It is not possible to access any of Aegis' HandlerDependencies in standard middleware
	// even if we were to pass them here because it's function signature simply doesn't consider
	// anything like that. It literally only deals with the request. Though the context will
	// be added to the request with Proxy().
	res, err := runStandardMiddleware(ctx, &req, r.stdMiddleware...)

	// Then run the Router middleware added with Use().
	if !runMiddleware(ctx, d, &req, &res, params, r.middleware...) {
		return res, nil
	}

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
		// Trace/capture the handler (in XRay by default) automatically
		r.Tracer.Record("annotation",
			map[string]interface{}{
				"RequestPath": req.Path,
				"Method":      req.HTTPMethod,
			},
		)

		err = r.Tracer.Capture(ctx, "RouteHandler", func(ctx1 context.Context) error {
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
