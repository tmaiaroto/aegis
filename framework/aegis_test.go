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
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAegis(t *testing.T) {

	// Mock API Gateway Router handler and Aegis interface
	handlers := Handlers{
		Router: NewRouter(func(ctx context.Context, d *HandlerDependencies, req *APIGatewayProxyRequest, res *APIGatewayProxyResponse, params url.Values) error {
			res.JSON(200, req)
			return nil
		}),
	}
	a := New(handlers)
	// Mock context and event
	ctx := context.Background()
	evt := map[string]interface{}{
		"httpMethod": "GET",
		"path":       "/",
		"stageVariables": map[string]string{
			"foo": "bar",
			"b64": "aGVsbG8gd29ybGQ=",
		},
		"requestContext": map[string]interface{}{
			"stage": "dev",
		},
	}

	Convey("New()", t, func() {
		Convey("Should be return a new Aegis interface", func() {
			So(a, ShouldHaveSameTypeAs, &Aegis{})
		})
	})

	Convey("ConfigureLogger()", t, func() {
		var logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
		a.ConfigureLogger(logger)
		Convey("Should allow a `*logrus.Logger` to replace the default `Log`", func() {
			So(a.Log.Level.String(), ShouldEqual, "warning")
		})
	})

	Convey("ConfigureService()", t, func() {
		a.ConfigureService("myservice", func(ctx context.Context, evt map[string]interface{}) interface{} {
			return map[string]string{"config": "value"}
		})
		Convey("Should set configurations for Services", func() {
			So(a.Services.configurations, ShouldContainKey, "myservice")
		})
	})

	Convey("setAegisVariables()", t, func() {
		a.setAegisVariables(ctx, evt)
		Convey("Should set variables from API Gateway stage variables", func() {
			So(a.Variables, ShouldContainKey, "foo")
		})

		Convey("Should handle base64 encoded values from API Gateway stage variables", func() {
			So(a.Variables, ShouldContainKey, "b64")
			So(a.Variables["b64"], ShouldEqual, "hello world")
		})
	})

	Convey("aegisHandler()", t, func() {
		a.Filters.Handler.BeforeServices = []func(*context.Context, map[string]interface{}){
			func(ctx *context.Context, evt map[string]interface{}) {
				evt["queryStringParameters"] = map[string]string{"alter": "event"}
			},
		}

		a.Filters.Handler.Before = []func(*context.Context, map[string]interface{}){
			func(ctx *context.Context, evt map[string]interface{}) {
				evt["queryStringParameters"].(map[string]string)["another"] = "alteration"
			},
		}

		a.Filters.Handler.After = []func(*context.Context, interface{}, error) (interface{}, error){
			func(ctx *context.Context, res interface{}, err error) (interface{}, error) {
				// You need to know something about the response here in order to apply a filter to it
				// In this case, it's an API Gateway Proxy Response so it's asserted: res.(APIGatewayProxyResponse)

				var apiGWResJSON map[string]interface{}
				if err := json.Unmarshal([]byte(res.(APIGatewayProxyResponse).Body), &apiGWResJSON); err != nil {
					panic(err)
				}
				apiGWResJSON["alteredResponse"] = "test"

				// A new interface{} and error get returned, overwriting the previous
				newRes := res.(APIGatewayProxyResponse)
				newRes.JSON(200, apiGWResJSON)
				return newRes, err
			},
		}

		res, err := a.aegisHandler(ctx, evt)
		apiGWRes := res.(APIGatewayProxyResponse)
		// The json decoded response to assert against
		var jsonRes map[string]interface{}
		if err := json.Unmarshal([]byte(apiGWRes.Body), &jsonRes); err != nil {
			panic(err)
		}

		Convey("Should handle a Lambda event", func() {
			// So(apiGWRes.Body, ShouldEqual, "Success")
			So(apiGWRes.StatusCode, ShouldEqual, 200)
			So(err, ShouldBeNil)
		})

		Convey("Should apply `BeforeServices` filters", func() {
			So(jsonRes["queryStringParameters"].(map[string]interface{}), ShouldContainKey, "alter")
		})

		Convey("Should apply `Before` filters", func() {
			So(jsonRes["queryStringParameters"].(map[string]interface{}), ShouldContainKey, "another")
		})

		Convey("Should apply `After` filters", func() {
			So(jsonRes, ShouldContainKey, "alteredResponse")
		})
	})

	Convey("requestToProxyRequest()", t, func() {
		Convey("Should take an HTTP request and format a Lambda Event", func() {
			localHandler := standAloneHandler{}
			r := httptest.NewRequest("GET", "/?foo=bar", strings.NewReader("some body to be read"))
			r.Header.Set("User-Agent", "aegis-test")

			_, req := localHandler.requestToProxyRequest(r)

			So(req.Body, ShouldEqual, "some body to be read")
			So(req.Headers, ShouldContainKey, "User-Agent")
			So(req.QueryStringParameters, ShouldContainKey, "foo")

		})
	})

	Convey("proxyResponseToHTTPResponse()", t, func() {
		Convey("Should take a Lambda Proxy response and format an HTTP response", func() {
			localHandler := standAloneHandler{}
			res := APIGatewayProxyResponse{
				StatusCode: 200,
				Headers:    map[string]string{"Content-Type": "application/json"},
			}
			rw := httptest.NewRecorder()
			localHandler.proxyResponseToHTTPResponse(&res, nil, rw)

			result := rw.Result()
			rw.Flush()
			So(result.StatusCode, ShouldEqual, 200)
			So(result.Header.Get("Content-Type"), ShouldEqual, "application/json")
		})
	})
}
