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
	"errors"
	"net/url"
)

// ValidAccessTokenMiddleware is helper middleware to verify a JWT from an `acess_token` cookie.
// It makes no determinations based on claims, it just looks for a valid token. A configured CognitoAppClient must be provided.
func ValidAccessTokenMiddleware(ctx context.Context, d *HandlerDependencies, req *APIGatewayProxyRequest, res *APIGatewayProxyResponse, params url.Values) bool {
	if d == nil || d.Services == nil || d.Services.Cognito == nil || d.Services.Cognito.ClientID == "" {
		res.JSONError(500, errors.New("auth has not been configured"))
		return false
	}

	jwtCookie, err := req.Cookie("access_token")
	if err != nil {
		res.JSONError(401, err)
		return false
	}
	// Check req Host and Referrer for increased protection

	// So we can use svc.Cognito now. Perhaps even check it's configured, right?
	// if svc != nil && svc.Cognito != nil
	_, err = d.Services.Cognito.ParseAndVerifyJWT(jwtCookie.Value)
	if err == nil {
		return true
	}

	res.JSONError(401, errors.New("unauthorized"))
	return false
}
