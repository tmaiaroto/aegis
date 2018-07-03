# Cognito

```go
router.Handle("GET", "/protected", cognitoProtected, aegis.ValidAccessTokenMiddleware)
```

There's a Cognito client helper interface which is obviously a lot more involved than a simple function.
Then there's also a smaller middleware helper function called <span class="nowrap">`ValidAccessTokenMiddleware`.</span>

You can use this middleware like any other middleware with your route handlers. It uses the Cognito
client interface's <span class="nowrap">`ParseAndVerifyJWT()`</span> function to check a JWT that you'll need to send
in under an `access_token` cookie in your HTTP request.

You can certainly check out <span class="nowrap">`cognito_helpers.go`</span> for how that works if you'd like to do
something similar on your own.