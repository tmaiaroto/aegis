// Copyright © 2016 Tom Maiaroto <tom@SerifAndSemaphore.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package framework

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	. "github.com/smartystreets/goconvey/convey"
)

type testProxyHandler struct{}

func (h testProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, you've hit %s", r.URL.Path)
}

func TestHandlerProxy(t *testing.T) {

	testHandler := func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "hello world")
	}
	adapter := NewHandlerAdapter(testHandler)

	Convey("NewHandlerAdapter()", t, func() {
		Convey("Should be return a new http handler adapter interface", func() {
			So(adapter, ShouldHaveSameTypeAs, &HandlerFuncAdapter{})
		})
	})

	Convey("Proxy()", t, func() {
		ctx := context.Background()

		Convey("Should return an events.APIGatewayProxyResponse", func() {
			evt := events.APIGatewayProxyRequest{
				Path: "/",
				Body: "",
			}
			res, _ := adapter.Proxy(ctx, evt)

			So(res, ShouldHaveSameTypeAs, events.APIGatewayProxyResponse{})
			So(res.Body, ShouldEqual, "hello world")
		})

		// Makes test output noisy. Not really sure the benefit of the test either.
		// Convey("Should return an error given an invalid event", func() {
		// 	evt := events.APIGatewayProxyRequest{
		// 		HTTPMethod: "bad method",
		// 		Path:       "/",
		// 		Body:       "",
		// 	}
		// 	So(func() { adapter.Proxy(ctx, evt) }, ShouldPanic)
		// })

		Convey("Should also work with an http.Handler", func() {
			adapter.Handler = testProxyHandler{}
			evt := events.APIGatewayProxyRequest{
				Path: "/",
				Body: "",
			}
			res, _ := adapter.Proxy(ctx, evt)
			So(res.Body, ShouldEqual, "hello, you've hit /")

		})

	})

}
