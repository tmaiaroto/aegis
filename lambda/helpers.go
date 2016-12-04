package lambda

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

const (
	charsetUTF8 = "charset=utf-8"
)

// Headers
const (
	HeaderAcceptEncoding                = "Accept-Encoding"
	HeaderAllow                         = "Allow"
	HeaderAuthorization                 = "Authorization"
	HeaderContentDisposition            = "Content-Disposition"
	HeaderContentEncoding               = "Content-Encoding"
	HeaderContentLength                 = "Content-Length"
	HeaderContentType                   = "Content-Type"
	HeaderCookie                        = "Cookie"
	HeaderSetCookie                     = "Set-Cookie"
	HeaderIfModifiedSince               = "If-Modified-Since"
	HeaderLastModified                  = "Last-Modified"
	HeaderLocation                      = "Location"
	HeaderUpgrade                       = "Upgrade"
	HeaderVary                          = "Vary"
	HeaderWWWAuthenticate               = "WWW-Authenticate"
	HeaderXForwardedProto               = "X-Forwarded-Proto"
	HeaderXHTTPMethodOverride           = "X-HTTP-Method-Override"
	HeaderXForwardedFor                 = "X-Forwarded-For"
	HeaderXRealIP                       = "X-Real-IP"
	HeaderServer                        = "Server"
	HeaderOrigin                        = "Origin"
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
)

// JSON sends a JSON response with status code.
func (res *ProxyResponse) JSON(status int, body interface{}) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMEApplicationJSONCharsetUTF8)

	data, err := json.Marshal(body)
	if err == nil {
		res.Body = string(data)
	}
}

// JSONP sends a JSONP response with status code.
func (res *ProxyResponse) JSONP(status int, callback string, body interface{}) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMEApplicationJSONCharsetUTF8)

	data, err := json.Marshal(body)
	if err == nil {
		var buffer bytes.Buffer
		buffer.WriteString(callback)
		buffer.WriteString("(")
		buffer.WriteString(string(data))
		buffer.WriteString(")")
		res.Body = buffer.String()
		buffer.Reset()
	}
}

// XML sends an XML response with status code.
func (res *ProxyResponse) XML(status int, i interface{}) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMEApplicationXMLCharsetUTF8)

	b, err := xml.Marshal(i)
	if err != nil {
		res.err = err
	} else {
		var buffer bytes.Buffer
		buffer.WriteString(xml.Header)
		buffer.WriteString(string(b))
		res.Body = buffer.String()
		buffer.Reset()
	}
}

// HTML sets an HTML header and returns a response with status code.
func (res *ProxyResponse) HTML(status int, html string) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMETextHTMLCharsetUTF8)
	res.Body = html
}

// String returns a plain text response with status code.
func (res *ProxyResponse) String(status int, s string) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMETextPlainCharsetUTF8)
	res.Body = s
}

// Error returns a JSON message with an error and status code.
func (res *ProxyResponse) Error(status int, e error) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMEApplicationJSONCharsetUTF8)

	data, err := json.Marshal(map[string]string{"error": e.Error()})
	if err == nil {
		res.Body = string(data)
	}
}

// Redirect redirects the request with status code.
func (res *ProxyResponse) Redirect(status int, url string) {
	if status < http.StatusMultipleChoices || status > http.StatusTemporaryRedirect {
		res.err = errors.New("Invalid redirect status code")
	} else {
		res.SetHeader(HeaderLocation, url)
		res.SetStatus(status)
	}
}

// SetHeader will set a ProxyResponse header replacing any existing value.
func (res *ProxyResponse) SetHeader(key string, value string) {
	res.Headers[key] = value
}

// GetHeader will return the value for a given header key. If there are no values associated with the key, GetHeader returns "".
func (evt *Event) GetHeader(key string) string {
	value := ""
	for k, v := range evt.Headers {
		if key == k {
			value = v
		}
	}
	return value
}

// SetStatus will set the status code for the response.
func (res *ProxyResponse) SetStatus(status int) {
	res.StatusCode = strconv.Itoa(status)
}

// TODO: add more response helpers like echo
// File? Attachment?

// IP returns the visitor's IP address from the event struct.
func (evt *Event) IP() string {
	return evt.RequestContext.Identity.SourceIp
}

// UserAgent returns the visitor's browser agent.
func (evt *Event) UserAgent() string {
	return evt.RequestContext.Identity.UserAgent
}

// GetParam returns a querystring parameter given its key name or empty string if not set.
func (evt *Event) GetParam(key string) string {
	param := ""
	if val, ok := evt.QueryStringParameters[key]; ok {
		param = val
	}
	return param
}

// GetForm will return a Form struct from a form-data body if passed in the event.
func (evt *Event) GetForm() (map[string]interface{}, error) {
	formData := map[string]interface{}{}
	mediaType, params, err := mime.ParseMediaType(evt.GetHeader(HeaderContentType))
	if err == nil {
		if strings.HasPrefix(mediaType, "multipart/") {
			body := strings.NewReader(evt.Body.(string))
			mr := multipart.NewReader(body, params["boundary"])
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					return formData, nil
				}
				if err != nil {
					return formData, err
				}
				b, err := ioutil.ReadAll(p)
				if err != nil {
					return formData, err
				}
				formData[p.FormName()] = string(b)
			}
		}
	}

	return formData, err
}
