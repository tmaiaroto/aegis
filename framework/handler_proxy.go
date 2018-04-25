// From https://github.com/awslabs/aws-lambda-go-api-proxy
// Aegis expands upon it a bit to include `Handler` as well
// as to export `HandlerFunc` field on `HandlerFuncAdapter`.

package framework

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
)

// HandlerFuncAdapter is an interface for adapting and handling Lambda APIGatewayProxyRequests to standard http Requests
type HandlerFuncAdapter struct {
	core.RequestAccessor
	HandlerFunc http.HandlerFunc
	Handler     http.Handler
}

// NewHandlerAdapter creates a new stdlib http.HandlerFunc adapter allowing the use of standard http request handling
func NewHandlerAdapter(handlerFunc http.HandlerFunc) *HandlerFuncAdapter {
	return &HandlerFuncAdapter{
		HandlerFunc: handlerFunc,
	}
}

// Proxy will proxy Lambda APIGatewayProxyRequests through a standard http handler and return an APIGatewayProxyResponse
func (h *HandlerFuncAdapter) Proxy(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req, err := h.ProxyEventToHTTPRequest(event)
	// Set the context from the Lambda event on to the http.Request
	req.WithContext(ctx)
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	w := core.NewProxyResponseWriter()
	// a Handler could include middleware, ie. when using Alice in chains.
	// Apollo is a fork there that may eventually be useful too.
	// So the `Handler` is generic, it need not be the alice package.
	if h.Handler != nil {
		h.Handler.ServeHTTP(http.ResponseWriter(w), req)
	} else {
		h.HandlerFunc.ServeHTTP(http.ResponseWriter(w), req)
	}

	resp, err := w.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return resp, nil
}
