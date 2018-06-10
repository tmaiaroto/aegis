# API Gateway Proxy (HTTP)

There are quite a few helpers for working with API Gateway Proxy responses and requests. You can read about
all of them through Go documentation.

 * <a href="https://godoc.org/github.com/tmaiaroto/aegis/framework#APIGatewayProxyResponse" target="_blank">APIGatewayProxyResponse</a>
 * <a href="https://godoc.org/github.com/tmaiaroto/aegis/framework#APIGatewayProxyRequest" target="_blank">APIGatewayProxyRequest</a>

Though let's take a look at some of the more common ones here.

## APIGatewayProxyRequest

API Gateway requests are not HTTP requests. They come into Lambda as JSON messages. The AWS Lambda Go package,
and therefore Aegis, marshals these messages to structs. That's great, but it still doesn't give us an easy
way to work with the request as if though it were an actual HTTP request. If you wanted to get a cookie, it
would be rather annoying to be frank.

**GetHeader()**

```go
auth := req.GetHeader("Authorization")
cookiesString := req.GetHeader("Cookie")
```

This will return a given header's value by name. Pretty easy, but you should note that getting the `Cookie`
header isn't exactly going to be real useful by itself. You'll need to parse that string. What you'll get
is a string like <span class="nowrap">`yummy_cookie=choco; tasty_cookie=strawberry`</span>, which isn't
great to then extract the one you're after.

**Cookie() and Cookies()**

```go
cookie := req.Cookie("yummy_cookie")
allCookies := req.Cookies()
```

So this helper function does just that. It mimicks Go's standard `http` package. In fact, Go's standard
package was used to create an empty HTTP request with just the cookie header set. That then allows the
`Cookie()` function to be used. So when you call this function, the same exact function is called against
Go's standard library using this fake request that was created.

Obviously then, `Cookies()` calls the standard HTTP request's same name function. It returns `[]*Cookie`.

**GetBody(), GetJSONBody(), and GetForm()**

```go
bodyString, err := req.GetBody()
bodyMap, err := req.GetJSONBody()
formMap, err := req.GetForm()
```

These helper functions are used for getting the body content of an API Gateway request. One will give you
a string while the other will give you a map. It saves you from needing to decode some things.

`GetForm()` can be helpful if you made a multipart POST request. If will parse the form values and return
to you a <span class="nowrap">`map[string]interface{}`</span>.

**UserAgent()**

We don't actually need to parse headers or use `GetHeader()` just to get the <span class="nowrap">`User-Agent`</span>,
instead there's simply <span class="nowrap">`UserAgent()`</span>.

**IP()**

AWS API Gateway will set the requesting client's IP address as well. So just like getting the `User-Agent`,
you can simply get the IP address using `IP()`. Both of these values are under <span class="nowrap">`req.RequestContext.Identity`</span>.
So they aren't hard to get at, but you have to type and remember more.

## APIGatewayProxyResponse

There are some great helper functions on the response struct as well. Obviously, these are all about returning
data to the client. Things like the HTTP status code, body content, and more.

**String(), HTML(), JSON(), JSONP(), XML(), and XMLPretty()**

These are your basic body content helpers. They all take a `status` int value as their first argument. This
automatically sets the status rather than requiring you to use `SetStatus()` or assign the struct property
yourself. The second argument will then vary a little.

Some methods will take an `interface{}` while others, like `HTML()` and `String()` will simply take a `string`.
Those taking an interface are going to be doing a little bit of work of course.