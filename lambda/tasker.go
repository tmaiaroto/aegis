package lambda

import (
	"log"
	"net/url"
	"os"
)

// Tasker struct provides an interface to handle scheduled tasks
type Tasker struct {
	handler TaskHandler
}

// TaskerIn provides a std in interface for Tasker (can be set mainly for testing)
var TaskerIn = os.Stdin

// TaskerOut provides a std out interface for Tasker (can be set mainly for testing)
var TaskerOut = os.Stdout

// TaskHandler is similar to RouteHandler except there is no response
type TaskHandler func(*Context, *Event, url.Values)

// Listen will start a task listener which acts much like a router except that it handles scheduled task events instead
func (t *Tasker) Listen() {
	RunStream(func(ctx *Context, evt *Event) *ProxyResponse {
		// url.Values are typically used for qureystring parameters.
		// However, this router uses them for path params.
		// Querystring parameters can be picked up from the *Event though.
		params := url.Values{}

		t.handler(ctx, evt, params)

		// There's a response, technically, because RunStream() needs one. It's empty though.
		return NewProxyResponse(200, map[string]string{}, "", nil)

	}, TaskerIn, TaskerOut)
}

// NewTasker simply returns a new Tasker struct and behaves a bit like Router
func NewTasker() *Tasker {
	return &Tasker{}
}

// Handle will a task by name with a handler
func (t *Tasker) Handle(name string, handler TaskHandler) {
	if name == "" {

	}
	log.Println("Handling ", name)
	// r.tree.addNode(method, r.URIVersion+path, handler, middleware...)
}
