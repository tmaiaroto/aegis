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
	// Could do a check for an environment variable here or something. If set, StartServer(), else Start().
	// That way the same code can be deployed to AWS and used locally for development.
	//
	// This is the normal listener when running in AWS Lambda.
	// app.Start()
	//
	// CLI only (single run):
	// StartSingle() will take an --event flag when running the binary (or go run main.go) that defines
	// a path to a file with a JSON event. It has some configuration options (or CLI flags can be used)
	// and it returns the response or error message to the CLI. It can optionally hide any logging too.
	// That way only the response/error is shown. It always exits, so the following app.StartServer()
	// call will never be reached, if it's uncommented.
	app.StartSingle()
	//
	// Local API
	// StartServer() will start a local HTTP server that transforms regular HTTP requests to events that
	// conform to APIGatewayProxyRequest. It makes testing an API locally very easy. Options can be set
	// here to deal with CORS concerns and other headers can be set too. Technically speaking, one could
	// use this to run their API outside of Lambda and Start() when running in Lambda...But why use Aegis
	// at that point? Like StartSingle(), It's best used for local development, testing, CI/CD, etc.
	app.StartServer()
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
	d.Log.Println("Hello Aegis logger!")
	return true
}
