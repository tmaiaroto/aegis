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
	Convey("An APIGatewayProxyRequest struct", t, func() {
		f, _ := ioutil.ReadFile("./example_event.json")
		var testReq APIGatewayProxyRequest
		json.Unmarshal(f, &testReq)

		Convey("Should be able to return a UserAgent", func() {
			So(testReq.UserAgent(), ShouldEqual, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.98 Safari/537.36")
		})

		Convey("Should be able to return an IP address", func() {
			So(testReq.IP(), ShouldEqual, "71.189.200.100")
		})

		Convey("Should be able to return a specific header", func() {
			So(testReq.GetHeader("X-Amz-Cf-Id"), ShouldEqual, "ekjIPhCoWuazdKIji2IdEy4G9DG1AgwunkAbTDE_Me93l_kprnBQPr==")
		})

		Convey("Should be able to return a querystring param", func() {
			So(testReq.GetParam("foo"), ShouldEqual, "bar")
		})

		Convey("Should be able to parse and return multipart/form data", func() {
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
	})

	Convey("An APIGatewayProxyResponse struct", t, func() {
		resp := APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "text/plain"},
		}

		Convey("Should be able to return a string", func() {
			resp.String(200, "foo")
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Body, ShouldEqual, "foo")
		})

		Convey("Should be able to return JSON", func() {
			resp.JSON(200, "{\"foo\": \"bar\"}")
			So(resp.StatusCode, ShouldEqual, 200)
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "application/json")
		})

		Convey("Should be able to return JSONP", func() {
			resp.JSONP(200, "myfunc", "{\"foo\": \"bar\"}")
			So(resp.StatusCode, ShouldEqual, 200)
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "application/json")
			So(resp.Body, ShouldEqual, `myfunc("{\"foo\": \"bar\"}")`)
		})

		Convey("Should be able to return XML", func() {
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

		Convey("Should be able to return HTML", func() {
			resp.HTML(200, "<html><body></body></html>")
			So(resp.StatusCode, ShouldEqual, 200)
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "text/html")
		})

		Convey("Should be able to return an HTTP redirect", func() {
			redirectURL := "http://google.com"
			resp.Redirect(301, redirectURL)
			So(resp.StatusCode, ShouldEqual, 301)
			So(resp.Headers["Location"], ShouldEqual, redirectURL)
		})

		Convey("Should be able to return a Go error string in a JSON response", func() {
			e := errors.New("hey, something went wrong")
			resp.Error(500, e)
			So(resp.StatusCode, ShouldEqual, 500)
			So(resp.Body, ShouldEqual, "{\"error\":\"hey, something went wrong\"}")
		})
	})

}
