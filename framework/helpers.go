package framework

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
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
func (res *APIGatewayProxyResponse) JSON(status int, body interface{}) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMEApplicationJSONCharsetUTF8)

	data, err := json.Marshal(body)
	if err == nil {
		res.Body = string(data)
	}
}

// JSONP sends a JSONP response with status code.
func (res *APIGatewayProxyResponse) JSONP(status int, callback string, body interface{}) {
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
func (res *APIGatewayProxyResponse) XML(status int, i interface{}) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMEApplicationXMLCharsetUTF8)

	b, err := xml.Marshal(i)
	if err != nil {
		// TODO: Figure out what to do with error now that it's not on struct
		// res.err = err
	} else {
		res.Body = formatXML(b)
	}
}

// XMLPretty sends an indented XML response with status code.
func (res *APIGatewayProxyResponse) XMLPretty(status int, i interface{}, indent string) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMEApplicationXMLCharsetUTF8)

	b, err := xml.MarshalIndent(i, "", indent)
	if err != nil {
		// TODO: Figure out what to do with error now that it's not on struct
		// res.err = err
	} else {
		res.Body = formatXML(b)
	}
}

// formatXML joins together an XML header and marshalled XML bytes for a completely formatted response
func formatXML(b []byte) string {
	var buffer bytes.Buffer
	buffer.WriteString(xml.Header)
	buffer.WriteString(string(b))
	xmlString := buffer.String()
	buffer.Reset()
	return xmlString
}

// HTML sets an HTML header and returns a response with status code.
func (res *APIGatewayProxyResponse) HTML(status int, html string) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMETextHTMLCharsetUTF8)
	res.Body = html
}

// String returns a plain text response with status code.
func (res *APIGatewayProxyResponse) String(status int, s string) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMETextPlainCharsetUTF8)
	res.Body = s
}

// Error returns a JSON message with an error and status code.
func (res *APIGatewayProxyResponse) Error(status int, e error) {
	res.SetStatus(status)
	res.SetHeader(HeaderContentType, MIMEApplicationJSONCharsetUTF8)

	data, err := json.Marshal(map[string]string{"error": e.Error()})
	if err == nil {
		res.Body = string(data)
	}
}

// Redirect redirects the request with status code.
func (res *APIGatewayProxyResponse) Redirect(status int, url string) {
	if status < http.StatusMultipleChoices || status > http.StatusTemporaryRedirect {
		// TODO: Figure out how to handle error now that it's not on the struct
		//res.err = errors.New("Invalid redirect status code")
	} else {
		res.SetHeader(HeaderLocation, url)
		res.SetStatus(status)
	}
}

// SetHeader will set a APIGatewayProxyResponse header replacing any existing value.
func (res *APIGatewayProxyResponse) SetHeader(key string, value string) {
	if res.Headers == nil {
		res.Headers = make(map[string]string)
	}
	res.Headers[key] = value
}

// SetStatus will set the status code for the response.
func (res *APIGatewayProxyResponse) SetStatus(status int) {
	res.StatusCode = status
}

// TODO: add more response helpers like echo
// File? Attachment?

// GetHeader will return the value for a given header key. If there are no values associated with the key, GetHeader returns "".
func (req *APIGatewayProxyRequest) GetHeader(key string) string {
	value := ""
	for k, v := range req.Headers {
		if key == k {
			value = v
		}
	}
	return value
}

// IP returns the visitor's IP address from the request event struct.
func (req *APIGatewayProxyRequest) IP() string {
	return req.RequestContext.Identity.SourceIP
}

// UserAgent returns the visitor's browser agent.
func (req *APIGatewayProxyRequest) UserAgent() string {
	return req.RequestContext.Identity.UserAgent
}

// GetParam returns a querystring parameter given its key name or empty string if not set.
func (req *APIGatewayProxyRequest) GetParam(key string) string {
	param := ""
	if val, ok := req.QueryStringParameters[key]; ok {
		param = val
	}
	return param
}

// GetForm will return a Form struct from a form-data body if passed in the request event.
func (req *APIGatewayProxyRequest) GetForm() (map[string]interface{}, error) {
	formData := map[string]interface{}{}
	mediaType, params, err := mime.ParseMediaType(req.GetHeader(HeaderContentType))
	if err == nil {
		if strings.HasPrefix(mediaType, "multipart/") {
			body := strings.NewReader(req.Body)
			mr := multipart.NewReader(body, params["boundary"])
			for {
				p, readerErr := mr.NextPart()
				if readerErr == io.EOF {
					return formData, nil
				}
				if readerErr != nil {
					return formData, readerErr
				}
				b, readPartsErr := ioutil.ReadAll(p)
				// Not bothering with the error here, it shouldn't really ever occur
				// if it were to, I think it'd be a buffer overflow...but that couldn't
				// happen because AWS Lambda has a POST limit that's far below that.
				// Other read errors would have already been seen by this point (above).
				if readPartsErr == nil {
					formData[p.FormName()] = string(b)
				}
			}
		}
	}

	return formData, err
}

// GetBody will return the request body if passed in the event. It's base64 encoded.
func (req *APIGatewayProxyRequest) GetBody() (string, error) {
	s := ""
	b, err := base64.StdEncoding.DecodeString(req.Body)
	if err == nil {
		s = string(b[:])
	}
	return s, err
}

// GetJSONBody will return the request body as map if passed in the event as a JSON string (which would be base64 encoded).
func (req *APIGatewayProxyRequest) GetJSONBody() (map[string]interface{}, error) {
	var m map[string]interface{}
	b, err := base64.StdEncoding.DecodeString(req.Body)
	if err == nil {
		err = json.Unmarshal([]byte(b), &m)
	}
	return m, err
}
