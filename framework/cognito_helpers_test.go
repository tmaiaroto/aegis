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
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCognitoHelpers(t *testing.T) {

	Convey("ValidAccessTokenMiddleware()", t, func() {

		Convey("Should be return boolean for valid access token", func() {
			d := &HandlerDependencies{}
			req := &APIGatewayProxyRequest{}
			res := &APIGatewayProxyResponse{}
			params := url.Values{}

			valid := ValidAccessTokenMiddleware(context.Background(), d, req, res, params)
			So(valid, ShouldBeFalse)
			So(res.StatusCode, ShouldEqual, 500)

			d.Services = &Services{
				Cognito: &CognitoAppClient{
					ClientID: "foo",
				},
			}
			valid = ValidAccessTokenMiddleware(context.Background(), d, req, res, params)
			So(valid, ShouldBeFalse)
			So(res.StatusCode, ShouldEqual, 401)
		})
	})

}
