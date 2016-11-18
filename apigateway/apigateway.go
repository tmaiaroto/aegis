// Copyright © 2016 Tom Maiaroto <tom@shift8creative.com>
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
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"
)

type Swagger struct {
	Swagger  string             `json:"swagger"`
	Info     APIInfo            `json:"info"`
	Host     string             `json:"host"`
	BasePath string             `json:"basePath"`
	Schemes  []string           `json:"schemes"`
	Paths    map[string]APIPath `json:"paths"`
}

type APIInfo struct {
	Version string `json:"version"`
	Title   string `json:"title"`
}

type APIPath struct {
	XAmazonApiGatwayAnyMethod ANYMethod `json:"x-amazon-apigateway-any-method"`
}

type ANYMethod struct {
	Produces                     []string                 `json:"produces"`
	Parameters                   []map[string]interface{} `json:"parameters"`
	Responses                    map[string]string        `json:"responses"`
	XAmazonApiGatewayIntegration APIIntegration           `json:"x-amazon-apigateway-integration"`
}

type APIIntegration struct {
	URI                 string                       `json:"uri"`
	Responses           map[string]map[string]string `json:"responses"`
	PassthroughBehavior string                       `json:"passthroughBehavior"`
	HttpMethod          string                       `json:"httpMethod"`
	CacheNamespace      string                       `json:"cacheNamespace"`
	CacheKeyParameters  []string                     `json:"cacheKeyParameters"`
	Type                string                       `json:"type"`
}

// SwaggerConfig holds configuration values for NewSwagger()
type SwaggerConfig struct {
	Title          string
	LambdaURI      string
	CacheNamespace string
	Version        string
}

func NewSwagger(cfg *SwaggerConfig) Swagger {
	if cfg.LambdaURI == "" {
		fmt.Println("Invalid Lambda URI provided for Swagger definition.")
		os.Exit(-1)
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
		XAmazonApiGatewayIntegration: APIIntegration{
			URI: cfg.LambdaURI,
			Responses: map[string]map[string]string{
				"default": map[string]string{
					"statusCode": "200",
				},
			},
			PassthroughBehavior: "when_no_match",
			HttpMethod:          "POST",
			CacheNamespace:      cfg.CacheNamespace,
			CacheKeyParameters:  []string{"method.request.path.proxy"},
			Type:                "aws_proxy",
		},
	}

	rootAnyMethod := ANYMethod{
		Produces:   []string{"application/json"},
		Parameters: params,
		Responses:  map[string]string{},
		XAmazonApiGatewayIntegration: APIIntegration{
			URI: cfg.LambdaURI,
			Responses: map[string]map[string]string{
				"default": map[string]string{
					"statusCode": "200",
				},
			},
			PassthroughBehavior: "when_no_match",
			HttpMethod:          "POST",
			Type:                "aws_proxy",
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
				XAmazonApiGatwayAnyMethod: rootAnyMethod,
			},
			"/{proxy+}": APIPath{
				XAmazonApiGatwayAnyMethod: proxyAnyMethod,
			},
		},
	}
}

// GetLambdaUri returns the Lambda URI
func GetLambdaUri(lambdaArn string) string {
	// lambdaArn won't work. It needs to be this format.
	// arn:aws:apigateway:<aws-region>:lambda:path/2015-03-31/functions/arn:aws:lambda:<aws-region>:<aws-acct-id>:function:<your-lambda-function-name>/invocations
	// ...but lambdaArn is in this string.

	// arn:aws:lambda:us-east-1:12345:function:aegis_example:6
	r, _ := regexp.Compile("arn:aws:lambda:(.+):([0-9]+):function:(.+)($|:)")
	matches := r.FindStringSubmatch(lambdaArn)
	accountId := ""
	region := ""
	functionName := ""
	if len(matches) == 5 {
		region = matches[1]
		accountId = matches[2]
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
	buffer.WriteString(accountId)
	buffer.WriteString(":function:")
	//buffer.WriteString(cfg.Lambda.FunctionName)
	buffer.WriteString(functionName)

	// Might not be able to use this...
	// buffer.WriteString(lambdaArn)
	buffer.WriteString("/invocations")
	lambdaUri := buffer.String()
	buffer.Reset()

	return lambdaUri
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

// Example Swagger Export for reference.
// {
//   "swagger": "2.0",
//   "info": {
//     "version": "2016-10-30T21:15:01Z",
//     "title": "Aegis"
//   },
//   "host": "xwvpe8m55b.execute-api.us-east-1.amazonaws.com",
//   "basePath": "/prod",
//   "schemes": [
//     "https"
//   ],
//   "paths": {
//     "/": {
//       "x-amazon-apigateway-any-method": {
//         "produces": [
//           "application/json"
//         ],
//         "responses": {},
//         "x-amazon-apigateway-integration": {
//           "uri": "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:12345:function:aegis_example/invocations",
//           "responses": {
//             "default": {
//               "statusCode": "200"
//             }
//           },
//           "passthroughBehavior": "when_no_match",
//           "httpMethod": "POST",
//           "type": "aws_proxy"
//         }
//       }
//     },
//     "/{proxy+}": {
//       "x-amazon-apigateway-any-method": {
//         "produces": [
//           "application/json"
//         ],
//         "parameters": [
//           {
//             "name": "proxy",
//             "in": "path",
//             "required": true,
//             "type": "string"
//           }
//         ],
//         "responses": {},
//         "x-amazon-apigateway-integration": {
//           "uri": "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:12345:function:aegistest/invocations",
//           "responses": {
//             "default": {
//               "statusCode": "200"
//             }
//           },
//           "passthroughBehavior": "when_no_match",
//           "httpMethod": "POST",
//           "cacheNamespace": "xxvfsk",
//           "cacheKeyParameters": [
//             "method.request.path.proxy"
//           ],
//           "type": "aws_proxy"
//         }
//       }
//     }
//   }
// }
