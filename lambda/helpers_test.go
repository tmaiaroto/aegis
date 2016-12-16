package lambda

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"strconv"
	"testing"
)

func TestHelpers(t *testing.T) {
	Convey("An Event struct", t, func() {
		f, _ := ioutil.ReadFile("./example_event.json")
		var testEvt Event
		json.Unmarshal(f, &testEvt)

		Convey("Should be able to return a UserAgent", func() {
			So(testEvt.UserAgent(), ShouldEqual, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.98 Safari/537.36")
		})

		Convey("Should be able to return an IP address", func() {
			So(testEvt.IP(), ShouldEqual, "71.189.200.100")
		})

		Convey("Should be able to return a specific header", func() {
			So(testEvt.GetHeader("X-Amz-Cf-Id"), ShouldEqual, "ekjIPhCoWuazdKIji2IdEy4G9DG1AgwunkAbTDE_Me93l_kprnBQPr==")
		})

		Convey("Should be able to return a querystring param", func() {
			So(testEvt.GetParam("foo"), ShouldEqual, "bar")
		})

		Convey("Should be able to parse and return multipart/form data", func() {
			testEvt.Headers["Content-Type"] = "multipart/form-data; boundary=------------------------ffdd24187066517d"
			testEvt.HTTPMethod = "POST"
			testEvt.Body = "\r\n--------------------------ffdd24187066517d\r\nContent-Disposition: form-data; name=\"text\"\r\n\r\ntext default\r\n--------------------------ffdd24187066517d--\r\n"
			testEvt.Headers["Content-Length"] = strconv.FormatInt(int64(len(testEvt.Body.(string))), 10)

			formData, err := testEvt.GetForm()
			So(formData, ShouldContainKey, "text")
			So(formData["text"], ShouldEqual, "text default")
			So(err, ShouldBeNil)

			testEvt.Body = "bad data"
			badFormData, err := testEvt.GetForm()
			So(badFormData, ShouldHaveSameTypeAs, map[string]interface{}{})
			So(err, ShouldNotBeNil)

			testEvt.Headers["Content-Type"] = "not multipart"
			noMultipartHeader, err := testEvt.GetForm()
			So(noMultipartHeader, ShouldHaveSameTypeAs, map[string]interface{}{})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("A ProxyResponse struct", t, func() {
		resp := NewProxyResponse(200, map[string]string{"Content-Type": "text/plain"}, "", nil)

		Convey("Should be able to return a string", func() {
			resp.String(200, "foo")
			So(resp.StatusCode, ShouldEqual, "200")
			So(resp.Body, ShouldEqual, "foo")
		})

		Convey("Should be able to return JSON", func() {
			resp.JSON(200, "{\"foo\": \"bar\"}")
			So(resp.StatusCode, ShouldEqual, "200")
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "application/json")
		})

		Convey("Should be able to return JSONP", func() {
			resp.JSONP(200, "myfunc", "{\"foo\": \"bar\"}")
			So(resp.StatusCode, ShouldEqual, "200")
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
			So(resp.StatusCode, ShouldEqual, "200")
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "application/xml")
			So(resp.Body, ShouldEqual, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<person><name>Tom</name></person>")

			resp.XMLPretty(200, xmlStruct, "    ")
			So(resp.StatusCode, ShouldEqual, "200")
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "application/xml")

			resp.XML(200, struct{ Foo string }{Foo: "string"})
			So(resp.err, ShouldNotBeNil)
			resp.XMLPretty(200, struct{ Foo string }{Foo: "string"}, "")
			So(resp.err, ShouldNotBeNil)
		})

		Convey("Should be able to return HTML", func() {
			resp.HTML(200, "<html><body></body></html>")
			So(resp.StatusCode, ShouldEqual, "200")
			// ShouldStarWith because charset is optional
			So(resp.Headers["Content-Type"], ShouldStartWith, "text/html")
		})

		Convey("Should be able to return an HTTP redirect", func() {
			redirectURL := "http://google.com"
			resp.Redirect(301, redirectURL)
			So(resp.StatusCode, ShouldEqual, "301")
			So(resp.Headers["Location"], ShouldEqual, redirectURL)

			resp.Redirect(200, redirectURL)
			So(resp.err, ShouldNotBeNil)
		})

		Convey("Should be able to return a Go error string in a JSON response", func() {
			e := errors.New("Hey, something went wrong.")
			resp.Error(500, e)
			So(resp.StatusCode, ShouldEqual, "500")
			So(resp.Body, ShouldEqual, "{\"error\":\"Hey, something went wrong.\"}")
		})
	})

}
