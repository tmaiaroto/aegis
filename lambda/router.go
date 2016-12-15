// Borrowed from https://github.com/acmacalister/helm

package lambda

import (
	"log"
	"net/url"
	"os"
	"strings"
)

const (
	get     = "GET"
	head    = "HEAD"
	post    = "POST"
	put     = "PUT"
	patch   = "PATCH"
	deleteh = "DELETE"
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
	r.Handle(head, path, handler, middleware...)
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
func (r *Router) PATCH(path string, handler RouteHandler, middleware ...Middleware) { // might make this and put one.
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
