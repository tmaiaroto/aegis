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
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHelpers(t *testing.T) {
	Convey("APIGatewayProxyRequest", t, func() {
		f, _ := ioutil.ReadFile("./example_event.json")
		var testReq APIGatewayProxyRequest
		json.Unmarshal(f, &testReq)

		Convey("UserAgent() Should be able to return a UserAgent", func() {
			So(testReq.UserAgent(), ShouldEqual, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.98 Safari/537.36")
		})

		Convey("IP() Should be able to return an IP address", func() {
			So(testReq.IP(), ShouldEqual, "71.189.200.100")
		})

		Convey("GetHeader() Should be able to return a specific header", func() {
			So(testReq.GetHeader("X-Amz-Cf-Id"), ShouldEqual, "ekjIPhCoWuazdKIji2IdEy4G9DG1AgwunkAbTDE_Me93l_kprnBQPr==")
		})

		Convey("GetParam() Should be able to return a querystring param", func() {
			So(testReq.GetParam("foo"), ShouldEqual, "bar")
		})

		Convey("GetForm() Should be able to parse and return multipart/form data", func() {
			testReq.Headers["Content-Type"] = "multipart/form-data; boundary=------------------------ffdd24187066517d"
			testReq.HTTPMethod = "POST"
			testReq.Body = "\r\n--------------------------ffdd24187066517d\r\nContent-Disposition: form-data; name=\"text\"\r\n\r\ntext default\r\n--------------------------ffdd24187066517d--\r\n"
			testReq.Headers["Content-Length"] = strconv.FormatInt(int64(len(testReq.Body)), 10)

			formData, err := testReq.GetForm()
			So(formData, ShouldContainKey, "text")
			So(formData["text"], ShouldEqual, "text default")
			So(err, ShouldBeNil)

			testReq.Body = "bad data"
			badFormData, err := testReq.GetForm()
			So(badFormData, ShouldHaveSameTypeAs, map[string]interface{}{})
			So(err, ShouldNotBeNil)

			testReq.Headers["Content-Type"] = "not multipart"
			noMultipartHeader, err := testReq.GetForm()
			So(noMultipartHeader, ShouldHaveSameTypeAs, map[string]interface{}{})
			So(err, ShouldNotBeNil)
		})

		Convey("Cookie() should be able to return an http.Cookie from a request", func() {
			cookie, err := testReq.Cookie("yummy_cookie")
			So(cookie.Name, ShouldEqual, "yummy_cookie")
			So(cookie.Value, ShouldEqual, "choco")
			So(err, ShouldBeNil)
		})

		Convey("Cookies() should be able to return all cookies from a request", func() {
			cookies, _ := testReq.Cookies()
			So(len(cookies), ShouldEqual, 2)
		})
	})

	Convey("APIGatewayProxyResponse", t, func() {
		resp := APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "text/plain"},
		}

		Convey("String() Should be able to return a response Body string", func() {
			resp.String(200, "foo")
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Body, ShouldEqual, "foo")
		})

		Convey("JSON() Should be able to return JSON", func() {
			resp.JSON(200, "{\"foo\": \"bar\"}")
			So(resp.StatusCode, ShouldEqual, 200)
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "application/json")
		})

		Convey("JSONP() Should be able to return JSONP", func() {
			resp.JSONP(200, "myfunc", "{\"foo\": \"bar\"}")
			So(resp.StatusCode, ShouldEqual, 200)
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "application/json")
			So(resp.Body, ShouldEqual, `myfunc("{\"foo\": \"bar\"}")`)
		})

		Convey("XML() and XMLPretty() Should be able to return XML", func() {
			xmlStruct := struct {
				XMLName xml.Name `xml:"person"`
				Name    string   `xml:"name"`
			}{
				Name: "Tom",
			}
			resp.XML(200, xmlStruct)
			So(resp.StatusCode, ShouldEqual, 200)
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "application/xml")
			So(resp.Body, ShouldEqual, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<person><name>Tom</name></person>")

			resp.XMLPretty(200, xmlStruct, "    ")
			So(resp.StatusCode, ShouldEqual, 200)
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "application/xml")
		})

		Convey("HTML() Should be able to return HTML", func() {
			resp.HTML(200, "<html><body></body></html>")
			So(resp.StatusCode, ShouldEqual, 200)
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "text/html")
		})

		Convey("Redirect() Should be able to return an HTTP redirect", func() {
			redirectURL := "http://google.com"
			resp.Redirect(301, redirectURL)
			So(resp.StatusCode, ShouldEqual, 301)
			So(resp.Headers["Location"], ShouldEqual, redirectURL)
		})

		Convey("Error() Should be able to return a Go error string", func() {
			e := errors.New("hey, something went wrong")
			resp.Error(500, e)
			So(resp.StatusCode, ShouldEqual, 500)
			So(resp.Body, ShouldEqual, "hey, something went wrong")
		})

		Convey("JSONError() Should be able to return a Go error string in a JSON response", func() {
			e := errors.New("hey, something went wrong")
			resp.JSONError(500, e)
			So(resp.StatusCode, ShouldEqual, 500)
			So(resp.Body, ShouldEqual, "{\"error\":\"hey, something went wrong\"}")
		})

		Convey("HTMLError() Should be able to return a Go error string in an HTML response", func() {
			e := errors.New("hey, something went wrong")
			resp.HTMLError(500, e)
			So(resp.StatusCode, ShouldEqual, 500)
			So(resp.Body, ShouldContainSubstring, "<html>")
			So(resp.Body, ShouldContainSubstring, "Error:")
		})

		Convey("XMLError() Should be able to return a Go error string in an XML response", func() {
			e := errors.New("hey, something went wrong")
			resp.XMLError(500, e)
			So(resp.StatusCode, ShouldEqual, 500)
			So(resp.Body, ShouldContainSubstring, "<?xml")
			So(resp.Body, ShouldContainSubstring, "<error>")
			So(resp.Body, ShouldContainSubstring, "<message>")
		})

		Convey("SetBodyError() should be able to set an error based on mime-type", func() {
			resp := APIGatewayProxyResponse{
				StatusCode: 200,
				Headers:    map[string]string{"Content-Type": "text/html"},
			}
			e := errors.New("hey, something went wrong")
			resp.SetBodyError(e)
			// Not set otherwise in this case, helper functions like JSONError() and Error() will automatically set the StatusCode
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Body, ShouldContainSubstring, "<html>")
			So(resp.Body, ShouldContainSubstring, "Error:")

		})
	})

}
