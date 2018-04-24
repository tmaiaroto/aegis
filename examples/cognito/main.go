package main

import (
	"context"
	"errors"
	"log"
	"net/url"

	aegis "github.com/tmaiaroto/aegis/framework"
)

func main() {
	router := aegis.NewRouter(fallThrough)
	router.Handle("GET", "/", root)
	router.Handle("GET", "/login", redirectToCognitoLogin)
	router.Handle("GET", "/logout", redirectToCognitoLogout)
	router.Handle("GET", "/callback", cognitoCallback)
	router.Handle("GET", "/protected", cognitoProtected, aegis.ValidAccessTokenMiddleware)

	AegisApp := aegis.New(aegis.Handlers{
		Router: router,
	})
	// The configuration is a function because config can come from a variety of sources
	AegisApp.ConfigureService("cognito", func(ctx context.Context, evt map[string]interface{}) interface{} {
		// Automatically get host. You don't need to do this - you typically would have a known domain name to use.
		host := ""
		stage := "prod"
		if headers, ok := evt["headers"]; ok {
			headersMap := headers.(map[string]interface{})
			if hostValue, ok := headersMap["Host"]; ok {
				host = hostValue.(string)
			}
		}
		// Automatically get API stage
		if requestContext, ok := evt["requestContext"]; ok {
			requestContextMap := requestContext.(map[string]interface{})
			if stageValue, ok := requestContextMap["stage"]; ok {
				stage = stageValue.(string)
			}
		}

		return &aegis.CognitoAppClientConfig{
			// The following three you'll need to fill in or use secrets.
			// See README for more info, but set secrets using aegis CLI and update aegis.yaml as needed.
			Region:   "us-east-1",
			PoolID:   AegisApp.GetVariable("PoolID"),
			ClientID: AegisApp.GetVariable("ClientID"),
			// This is just automatic for the example, you would likely replace this too
			RedirectURI:       "https://" + host + "/" + stage + "/callback",
			LogoutRedirectURI: "https://" + host + "/" + stage + "/",
		}
	})
	AegisApp.Start()
}

// fallThrough handles any path that couldn't be matched to another handler
func fallThrough(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	res.StatusCode = 404
	return nil
}

// root is handling GET "/" in this case
func root(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	host := req.GetHeader("Host")
	stage := req.RequestContext.Stage
	res.HTML(200, "<h1>Welcome</h1><p>This is an unprotected route.<br /><a href=\"https://"+host+"/"+stage+"/login\">Click here to login.</a></p>")
	return nil
}

// Redirect to AWS Cognito hosted login page
func redirectToCognitoLogin(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	res.Redirect(301, d.Services.Cognito.HostedLoginURL)
	return nil
}

// Redirect to AWS Cognito hosted logout page and remove our domain cookie
func redirectToCognitoLogout(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	host := req.GetHeader("Host")
	res.SetHeader("Set-Cookie", "access_token=; Domain="+host+"; Secure; HttpOnly")
	res.Redirect(301, d.Services.Cognito.HostedLogoutURL)
	return nil
}

// Handle oauth2 callback, will exchange code for token
func cognitoCallback(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	// Exchange code for token
	tokens, err := d.Services.Cognito.GetTokens(req.QueryStringParameters["code"], []string{})
	if err != nil {
		log.Println("Couldn't get access token", err)
		res.JSONError(500, err)
	} else {
		// verify the token
		_, err := d.Services.Cognito.ParseAndVerifyJWT(tokens.IDToken)
		if err == nil {
			host := req.GetHeader("Host")
			stage := req.RequestContext.Stage
			res.SetHeader("Set-Cookie", "access_token="+tokens.AccessToken+"; Domain="+host+"; Secure; HttpOnly")
			res.Redirect(301, "https://"+host+"/"+stage+"/protected")
		} else {
			res.JSONError(401, errors.New("unauthorized, invalid token"))
		}
	}
	return nil
}

// Example after successful login
func cognitoProtected(ctx context.Context, d *aegis.HandlerDependencies, req *aegis.APIGatewayProxyRequest, res *aegis.APIGatewayProxyResponse, params url.Values) error {
	host := req.GetHeader("Host")
	stage := req.RequestContext.Stage
	res.HTML(200, `<h1>Success!</h1>
		<p>You are logged in. This endpoint would not be accessible if you did not have a valid JWT.</p>
		<a href="https://`+host+`/`+stage+`/logout">Click here to logout.</a>
	`)
	return nil
}
