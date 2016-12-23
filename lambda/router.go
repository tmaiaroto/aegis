// Borrowed from https://github.com/acmacalister/helm

package lambda

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	get     = "GET"
	head    = "HEAD"
	post    = "POST"
	put     = "PUT"
	patch   = "PATCH"
	deleteh = "DELETE"
	options = "OPTIONS"
)

// RouteHandler is similar to "net/http" Handlers, except there is no response writer.
// Instead the *ProxyResponse is manipulated and returned directly.
type RouteHandler func(*Context, *Event, *ProxyResponse, url.Values)

// Middleware is just like the RouteHandler type, but has a boolean return. True
// means to keep processing the rest of the middleware chain, false means end.
type Middleware func(*Context, *Event, *ProxyResponse, url.Values) bool

// Router name says it all.
type Router struct {
	tree           *node
	rootHandler    RouteHandler
	middleware     []Middleware
	l              *log.Logger
	LoggingEnabled bool
	URIVersion     string
	GatewayPort    string
}

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
	r.Handle(deleteh, path, handler, middleware...)
}

// runMiddleware loops over the slice of middleware and call to each of the middleware handlers.
func runMiddleware(ctx *Context, evt *Event, res *ProxyResponse, params url.Values, middleware ...Middleware) bool {
	for _, m := range middleware {
		if !m(ctx, evt, res, params) {
			return false // the middleware returned false, so end processing the chain.
		}
	}
	return true
}

// Listen will start the internal router and listen for Lambda events to forward to registered routes.
func (r *Router) Listen() {
	RunStream(func(ctx *Context, evt *Event) *ProxyResponse {
		// url.Values are typically used for qureystring parameters.
		// However, this router uses them for path params.
		// Querystring parameters can be picked up from the *Event though.
		params := url.Values{}
		// New empty response with a 200 status code since nothing has gone wrong yet, it's just empty.
		res := NewProxyResponse(200, map[string]string{}, "", nil)

		// use the Path and HTTPMethod from the event to figure out the route
		node, _ := r.tree.traverse(strings.Split(evt.Path, "/")[1:], params)
		if handler := node.methods[evt.HTTPMethod]; handler != nil {
			// Middleware must return true in order to continue.
			// If it returns false, it will catch and halt everything.
			if !runMiddleware(ctx, evt, res, params, handler.middleware...) {
				// TODO: Figure out what to do here. I'm not sure what makes sense.
				// Should it return the response in its current state?
				return res
				// Or should it return an error?
				// Typically it leaves the request hanging if it returns false.
				// The middleware would need to write something back to the client.
				// return NewProxyResponse(500, map[string]string{}, "", nil)
			}
			handler.handler(ctx, evt, res, params)
		} else {
			r.rootHandler(ctx, evt, res, params)
		}

		return res
	}, os.Stdin, os.Stdout)
}

// gatewayHandler is a Router that implements an http.Handler interface
type gatewayHandler Router

// ServerHTTP will build an Event using an HTTP Request, run the appropriate handler from the Lambda code and then return an HTTP Response.
// Normally AWS API Gateway handles this for us, but locally, we'll need to do these transforms.
func (h gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// --> Build the request by converting HTTP request to Lambda Event
	ctx, evt := h.requestToEvent(r)

	// -<>- Normal handling of Lambda Event with Aegis Router.
	// It looks just like the closure given to RunStream with Listen(), except that we aren't going
	// to deal with stdio. Instead, we're taking the results and passing back through an HTTP Response.
	// TODO: refactor this to make a new handler that also works with a router. There's a tiny bit of repetition with this and what's in Listen()
	params := url.Values{}
	// New empty response with a 200 status code since nothing has gone wrong yet, it's just empty.
	res := NewProxyResponse(200, map[string]string{}, "", nil)
	// use the Path and HTTPMethod from the event to figure out the route
	node, _ := h.tree.traverse(strings.Split(evt.Path, "/")[1:], params)
	if handler := node.methods[evt.HTTPMethod]; handler != nil {
		// Middleware must return true in order to continue.
		// If it returns false, it will catch and halt everything.
		if !runMiddleware(ctx, evt, res, params, handler.middleware...) {
			// <-- Send the response
			h.proxyResponseToHTTPResponse(res, w)
			return
		}
		handler.handler(ctx, evt, res, params)
	} else {
		h.rootHandler(ctx, evt, res, params)
	}

	// <-- Send the response
	h.proxyResponseToHTTPResponse(res, w)
}

// requestToEvent will take an HTTP Request and transform it into a faux AWS Lambda Event.
// It will be missing some data like AWS Cognito information, API stage information, and more.
// However, it will be enough to handle request/response and will mimic Lambda locally.
// AWS does this for us automatically, but when running a local HTTP server, we'll need to do it.
func (h gatewayHandler) requestToEvent(r *http.Request) (*Context, *Event) {
	ctx := Context{}
	evt := Event{
		Path:       r.URL.Path,
		HTTPMethod: r.Method,
		RequestContext: RequestContext{
			HTTPMethod: r.Method,
		},
		HandlerStartTimeMs: time.Now().Unix(),
		HandlerStartTime:   time.Now().UnixNano(),
	}
	// transfer the headers over to the event
	evt.Headers = map[string]string{}
	for k, v := range r.Header {
		evt.Headers[k] = strings.Join(v, "; ")
	}

	// Querystring params
	params := r.URL.Query()
	paramsMap := map[string]string{}
	for k, _ := range params {
		paramsMap[k] = params.Get(k)
	}
	evt.QueryStringParameters = paramsMap

	// Path params (just the proxy+ path ... but it does not have the preceding slash)
	evt.PathParameters = map[string]string{
		"proxy": r.URL.Path[1:len(r.URL.Path)],
	}
	evt.Resource = "/{proxy+}"
	evt.RequestContext.ResourcePath = "/{proxy+}"

	// Identity info: user agent, IP, etc.
	evt.RequestContext.Identity.UserAgent = r.Header.Get("User-Agent")
	evt.RequestContext.Identity.SourceIP = r.Header.Get("X-Forwarded-For")
	if evt.RequestContext.Identity.SourceIP == "" {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			evt.RequestContext.Identity.SourceIP = net.ParseIP(ip).String()
		}
	}

	// Stage will be "local" for now? I'm not sure what makes sense here. Local gateway. Local. Debug. ¯\_(ツ)_/¯
	evt.RequestContext.Stage = "local"
	// TODO: Stage variables would need to be pulled from the aegis.yaml ...
	// so now the config file has to be next to the app... otherwise some defaults will be set like "local"
	// and no stage variables i suppose.
	// evt.StageVariables =

	// The request id will simply be a timestamp to help keep it unique, but also allowing it to be easily sorted
	evt.RequestContext.RequestID = strconv.FormatInt(time.Now().UnixNano(), 10)

	// pass along the body
	bodyData, err := ioutil.ReadAll(r.Body)
	if err == nil {
		evt.Body = string(bodyData)
	}

	return &ctx, &evt
}

// proxyResponseToHTTPResponse will take the typical Lambda Proxy response and transform it into an HTTP response.
// AWS does this for us automatically, but when running a local HTTP server, we'll need to do it.
func (h gatewayHandler) proxyResponseToHTTPResponse(res *ProxyResponse, w http.ResponseWriter) {
	// transfer the headers into the HTTP Response
	for k, v := range res.Headers {
		w.Header().Set(k, v)
	}

	// The handler and middleware should have set everything on res
	code, _ := strconv.ParseInt(res.StatusCode, 10, 32)
	w.WriteHeader(int(code))

	// and of course the body
	fmt.Fprintf(w, res.Body)
}

// Gateway will start a local web server to listen for events. Useful for testing locally.
func (r *Router) Gateway() {
	if r.GatewayPort == "" {
		r.GatewayPort = ":9999"
	}
	fmt.Printf("Starting local gateway: http://localhost%v \n", r.GatewayPort)
	err := http.ListenAndServe(r.GatewayPort, gatewayHandler(*r))
	if err != nil {
		log.Fatal(err)
	}
}
