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

	// Use an Aegis interface to inject optional dependencies into handlers
	// and start listening for events.
	app := aegis.New(aegis.Handlers{
		Router: router,
	})
	app.Start()
}

// fallThrough handles any path that couldn't be matched to another handler
func fallThrough(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	res.StatusCode = 404
	return nil
}

// root is handling GET "/" in this case
func root(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	lc, _ := lambdacontext.FromContext(ctx)
	res.JSON(200, map[string]interface{}{"event": req, "context": lc})
	return nil
}

// helloMiddleware is a simple example of middleware
func helloMiddleware(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) bool {
	log.Println("Hello CloudWatch!")
	return true
}
