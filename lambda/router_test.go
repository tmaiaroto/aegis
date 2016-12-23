package lambda

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRouter(t *testing.T) {
	testFallThroughHandler := func(ctx *Context, evt *Event, res *ProxyResponse, params url.Values) {}
	testHandler := func(ctx *Context, evt *Event, res *ProxyResponse, params url.Values) {}
	testMiddleware := func(ctx *Context, evt *Event, res *ProxyResponse, params url.Values) bool { return true }
	testMiddlewareStop := func(ctx *Context, evt *Event, res *ProxyResponse, params url.Values) bool { return false }
	testParams := url.Values{}
	testRouter := NewRouter(testFallThroughHandler)
	Convey("NewRouter", t, func() {
		Convey("Should create a new Router", func() {
			So(testRouter, ShouldNotBeNil)
		})
	})

	testRouter.GET("/path", testHandler)
	testRouter.POST("/path", testHandler)
	testRouter.PUT("/path", testHandler)
	testRouter.PATCH("/path", testHandler)
	testRouter.DELETE("/path", testHandler)
	testRouter.HEAD("/path", testHandler)
	testRouter.OPTIONS("/path", testHandler)

	node, _ := testRouter.tree.traverse(strings.Split("/path", "/")[1:], testParams)

	Convey("Should handle GET", t, func() {
		So(node.methods, ShouldContainKey, "GET")
	})

	Convey("Should handle POST", t, func() {
		So(node.methods, ShouldContainKey, "POST")
	})

	Convey("Should handle PUT", t, func() {
		So(node.methods, ShouldContainKey, "PUT")
	})

	Convey("Should handle PATCH", t, func() {
		So(node.methods, ShouldContainKey, "PATCH")
	})

	Convey("Should handle DELETE", t, func() {
		So(node.methods, ShouldContainKey, "DELETE")
	})

	Convey("Should handle HEAD", t, func() {
		So(node.methods, ShouldContainKey, "HEAD")
	})

	Convey("Should handle OPTIONS", t, func() {
		So(node.methods, ShouldContainKey, "OPTIONS")
	})

	Convey("runMiddleware", t, func() {
		Convey("Should handle middleware", func() {
			ctx := Context{}
			evt := Event{}
			res := ProxyResponse{}
			params := url.Values{}
			next := runMiddleware(&ctx, &evt, &res, params, testMiddleware)
			So(next, ShouldBeTrue)

			noNext := runMiddleware(&ctx, &evt, &res, params, testMiddlewareStop, testMiddleware)
			So(noNext, ShouldBeFalse)
		})
	})

	Convey("requestToEvent", t, func() {
		Convey("Should take an HTTP request and format a Lambda Event", func() {
			gwHandler := gatewayHandler{}
			r := httptest.NewRequest("GET", "/?foo=bar", strings.NewReader("some body to be read"))
			r.Header.Set("User-Agent", "aegis-test")

			ctx, evt := gwHandler.requestToEvent(r)

			So(ctx.FunctionVersion, ShouldBeEmpty)
			So(evt.Body.(string), ShouldEqual, "some body to be read")
			So(evt.Headers, ShouldContainKey, "User-Agent")
			So(evt.QueryStringParameters, ShouldContainKey, "foo")

		})
	})

	Convey("proxyResponseToHTTPResponse", t, func() {
		Convey("Should take a Lambda Proxy response and format an HTTP response", func() {
			gwHandler := gatewayHandler{}
			res := NewProxyResponse(200, map[string]string{"Content-Type": "application/json"}, "", nil)
			rw := httptest.NewRecorder()
			gwHandler.proxyResponseToHTTPResponse(res, rw)

			result := rw.Result()
			rw.Flush()
			So(result.StatusCode, ShouldEqual, 200)
			So(result.Header.Get("Content-Type"), ShouldEqual, "application/json")
		})
	})
}
