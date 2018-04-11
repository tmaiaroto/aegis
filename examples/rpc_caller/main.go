package main

import (
	"context"
	"log"
	"net/url"

	"github.com/aws/aws-lambda-go/lambdacontext"
	aegis "github.com/tmaiaroto/aegis/framework"
)

func main() {
	// Handle an APIGatewayProxyRequest event with a URL reqeust path Router
	router := aegis.NewRouter(fallThrough)
	router.Handle("GET", "/", root, helloMiddleware)
	router.Listen()
}

// fallThrough handles any path that couldn't be matched to another handler
func fallThrough(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	res.StatusCode = 404
	return nil
}

// root is handling GET "/" in this case
func root(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	lc, _ := lambdacontext.FromContext(ctx)

	// RPC example
	// rpcPayload := map[string]interface{}{
	// 	"_rpcName": "procedure",
	// 	"foo":      "bar",
	// }
	// resp, rpcErr := aegis.RPC(ctx, "aegis_example", rpcPayload)
	rpcPayload := map[string]interface{}{
		"_rpcName":  "lookup",
		"ipAddress": req.RequestContext.Identity.SourceIP,
	}
	// resp, rpcErr := aegis.RPC(ctx, "aegis_geoip", rpcPayload)
	// Use Aegis interface's RPC() call to trace. This is an untraced Lambda invocation.
	// Which means it could be called outside of an event handler.
	resp, rpcErr := aegis.RPC("aegis_geoip", rpcPayload)
	log.Println(rpcErr)

	res.JSON(200, map[string]interface{}{"event": resp, "context": lc})
	return nil
}

// helloMiddleware is a simple example of middleware
func helloMiddleware(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) bool {
	log.Println("Hello CloudWatch!")
	return true
}
