package lambda

import (
	"log"
	"net/url"
)

const (
	get     = "GET"
	head    = "HEAD"
	post    = "POST"
	put     = "PUT"
	patch   = "PATCH"
	deleteh = "DELETE"
	any     = "ANY"
)

// RouteHandler is just like "net/http" Handlers, only takes params.
type RouteHandler func(string, string, url.Values)

// Middleware is just like the RouteHandler type, but has a boolean return. True
// means to keep processing the rest of the middleware chain, false means end.
// If you return false to end the request-response cycle you MUST
// write something back to the client, otherwise it will be left hanging.
type Middleware func(string, string, url.Values) bool

// Router name says it all.
type Router struct {
	tree           *node
	rootHandler    RouteHandler
	middleware     []Middleware
	l              *log.Logger
	LoggingEnabled bool
	URIVersion     string
}

type Param struct {
	Name     string
	Required bool
}

// New creates a new router. Take the root/fall through route
// like how the default mux works. Only difference is in this case,
// you have to specific one.
func NewRouter(rootHandler RouteHandler) *Router {
	node := node{component: "/", isNamedParam: false, methods: make(map[string]*route)}
	return &Router{tree: &node, rootHandler: rootHandler, URIVersion: ""}
}
