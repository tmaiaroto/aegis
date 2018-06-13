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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRPCRouter(t *testing.T) {

	fallThroughHandled := false
	rpcRouter := NewRPCRouter(func(ctx context.Context, d *HandlerDependencies, evt map[string]interface{}) (map[string]interface{}, error) {
		fallThroughHandled = true
		return map[string]interface{}{"fall": "through"}, nil
	})
	rpcRouter.Tracer = &NoTraceStrategy{}

	Convey("NewRPCRouter", t, func() {
		Convey("Should create a new NewRPCRouter", func() {
			So(rpcRouter, ShouldNotBeNil)
			So(rpcRouter, ShouldHaveSameTypeAs, &RPCRouter{})
		})
	})

	Convey("Should handle specific events with `_rpcName`", t, func() {
		handled := false
		rpcRouter.Handle("foo", func(ctx context.Context, d *HandlerDependencies, evt map[string]interface{}) (map[string]interface{}, error) {
			handled = true
			return map[string]interface{}{}, nil
		})

		So(handled, ShouldBeFalse)
		rpcRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, map[string]interface{}{"_rpcName": "foo"})
		So(handled, ShouldBeTrue)
	})

	Convey("Should optionally handle events with `_rpcName` using a fall through handler", t, func() {
		So(fallThroughHandled, ShouldBeFalse)
		rpcRouter.LambdaHandler(context.Background(), &HandlerDependencies{}, map[string]interface{}{"_rpcName": "unhandled"})
		So(fallThroughHandled, ShouldBeTrue)
	})

}
