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

// Package apigateway provides swagger for API Gateway.
package apigateway

import (
	"bytes"
	"errors"
	"math/rand"
	"regexp"
	"time"
)

// Swagger defines an AWS API Gateway Lambda Proxy swagger definition
// More info here: https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-swagger-extensions.html
type Swagger struct {
	Swagger                           string             `json:"swagger"`
	Info                              APIInfo            `json:"info"`
	Host                              string             `json:"host"`
	BasePath                          string             `json:"basePath"`
	Schemes                           []string           `json:"schemes"`
	Paths                             map[string]APIPath `json:"paths"`
	XAmazonAPIGatewayBinaryMediaTypes []string           `json:"x-amazon-apigateway-binary-media-types"`
}

// APIInfo provides is a part of th API Gateway swagger
type APIInfo struct {
	Version string `json:"version"`
	Title   string `json:"title"`
}

// APIPath provides is a part of th API Gateway swagger
type APIPath struct {
	XAmazonAPIGatwayAnyMethod ANYMethod `json:"x-amazon-apigateway-any-method"`
}

// ANYMethod provides is a part of th API Gateway swagger, it instructs the API to handle any HTTP method on an APIPath
type ANYMethod struct {
	Produces                     []string                 `json:"produces"`
	Parameters                   []map[string]interface{} `json:"parameters"`
	Responses                    map[string]string        `json:"responses"`
	XAmazonAPIGatewayIntegration APIIntegration           `json:"x-amazon-apigateway-integration"`
}

// APIIntegration provides is a part of th API Gateway swagger
type APIIntegration struct {
	URI                 string                       `json:"uri"`
	Responses           map[string]map[string]string `json:"responses"`
	PassthroughBehavior string                       `json:"passthroughBehavior"`
	HTTPMethod          string                       `json:"httpMethod"`
	CacheNamespace      string                       `json:"cacheNamespace"`
	CacheKeyParameters  []string                     `json:"cacheKeyParameters"`
	Type                string                       `json:"type"`
	TimeoutMs           int                          `json:"timeoutInMillis"`
}

// SwaggerConfig holds configuration values for NewSwagger()
// TODO: Probably want to a more comprehensive field for mapping XAmazonAPIGatewayIntegration values with defaults, etc.
// For now, "ResourceTimeoutMs" is really XAmazonAPIGatewayIntegration.TimeoutMs which is "timeoutInMillis" in Swagger JSON to AWS.
type SwaggerConfig struct {
	Title             string
	LambdaURI         string
	CacheNamespace    string
	Version           string
	ResourceTimeoutMs int
	// BinaryMediaTypes []string
}

// NewSwagger creates a new Swagger struct with some default values
func NewSwagger(cfg *SwaggerConfig) (Swagger, error) {
	if cfg.LambdaURI == "" {
		return Swagger{}, errors.New("invalid Lambda URI provided for Swagger definition")
	}

	// Some defaults
	if cfg.Title == "" {
		cfg.Title = "Example Aegis API"
	}

	if cfg.CacheNamespace == "" {
		cfg.CacheNamespace = randomCacheNamespace(6)
	}

	// Can version be anything but a timestamp?
	if cfg.Version == "" {
		t := time.Now()
		cfg.Version = t.UTC().Format(time.RFC3339)
	}

	// 29000 is the default timeout (29 seconds), let's keep that (instead of struct default to empty 0).
	if cfg.ResourceTimeoutMs == 0 {
		cfg.ResourceTimeoutMs = 29000
	}
	// 50 (50ms) is the lowest value possible.
	if cfg.ResourceTimeoutMs < 50 {
		cfg.ResourceTimeoutMs = 50
	}

	// Set the API Info
	apiInfo := APIInfo{
		Title:   cfg.Title,
		Version: cfg.Version,
	}

	// Set the ANY method
	params := []map[string]interface{}{}
	params = append(params, map[string]interface{}{
		"name":     "proxy",
		"in":       "path",
		"required": true,
		"type":     "string",
	})

	proxyAnyMethod := ANYMethod{
		Produces:   []string{"application/json"},
		Parameters: params,
		Responses:  map[string]string{},
		XAmazonAPIGatewayIntegration: APIIntegration{
			URI: cfg.LambdaURI,
			Responses: map[string]map[string]string{
				"default": map[string]string{
					"statusCode": "200",
				},
			},
			PassthroughBehavior: "when_no_match",
			HTTPMethod:          "POST",
			CacheNamespace:      cfg.CacheNamespace,
			CacheKeyParameters:  []string{"method.request.path.proxy"},
			Type:                "aws_proxy",
			TimeoutMs:           cfg.ResourceTimeoutMs,
		},
	}

	rootAnyMethod := ANYMethod{
		Produces:   []string{"application/json"},
		Parameters: params,
		Responses:  map[string]string{},
		XAmazonAPIGatewayIntegration: APIIntegration{
			URI: cfg.LambdaURI,
			Responses: map[string]map[string]string{
				"default": map[string]string{
					"statusCode": "200",
				},
			},
			PassthroughBehavior: "when_no_match",
			HTTPMethod:          "POST",
			Type:                "aws_proxy",
			TimeoutMs:           cfg.ResourceTimeoutMs,
		},
	}

	return Swagger{
		Swagger: "2.0",
		Info:    apiInfo,
		// Omit Host?
		BasePath: "/prod",
		Schemes:  []string{"https"},
		Paths: map[string]APIPath{
			"/": APIPath{
				XAmazonAPIGatwayAnyMethod: rootAnyMethod,
			},
			"/{proxy+}": APIPath{
				XAmazonAPIGatwayAnyMethod: proxyAnyMethod,
			},
		},
		// This does not work.
		// XAmazonAPIGatewayBinaryMediaTypes: cfg.BinaryMediaTypes,
	}, nil
}

// GetLambdaURI returns the Lambda URI
func GetLambdaURI(lambdaArn string) string {
	// lambdaArn won't work. It needs to be this format.
	// arn:aws:apigateway:<aws-region>:lambda:path/2015-03-31/functions/arn:aws:lambda:<aws-region>:<aws-acct-id>:function:<your-lambda-function-name>/invocations
	// ...but lambdaArn is in this string.

	// arn:aws:lambda:us-east-1:12345:function:aegis_example:6
	r := regexp.MustCompile("arn:aws:lambda:(.+):([0-9]+):function:(.+)($|:)")
	matches := r.FindStringSubmatch(lambdaArn)
	accountID := ""
	region := ""
	functionName := ""
	if len(matches) == 5 {
		region = matches[1]
		accountID = matches[2]
		functionName = matches[3]
	}

	// arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:12345:function:aegis_example:6/invocations
	// arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:12345:function:aegistest/invocations,
	var buffer bytes.Buffer
	buffer.WriteString("arn:aws:apigateway:")
	// buffer.WriteString(cfg.AWS.Region)
	buffer.WriteString(region)
	buffer.WriteString(":lambda:path/2015-03-31/functions/")
	buffer.WriteString("arn:aws:lambda:")
	//buffer.WriteString(cfg.AWS.Region)
	buffer.WriteString(region)
	buffer.WriteString(":")
	buffer.WriteString(accountID)
	buffer.WriteString(":function:")
	//buffer.WriteString(cfg.Lambda.FunctionName)
	buffer.WriteString(functionName)

	// Might not be able to use this...
	// buffer.WriteString(lambdaArn)
	buffer.WriteString("/invocations")
	lambdaURI := buffer.String()
	buffer.Reset()

	return lambdaURI
}

// randomCacheNamespace creates a random string to use for cache namespace
func randomCacheNamespace(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	var src = rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
