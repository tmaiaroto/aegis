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
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTrie(t *testing.T) {
	testFallThroughHandler := func(ctx context.Context, d *HandlerDependencies, req *APIGatewayProxyRequest, res *APIGatewayProxyResponse, params url.Values) error {
		return nil
	}
	testHandler := func(ctx context.Context, d *HandlerDependencies, req *APIGatewayProxyRequest, res *APIGatewayProxyResponse, params url.Values) error {
		return nil
	}
	testNamedHandler := func(ctx context.Context, d *HandlerDependencies, req *APIGatewayProxyRequest, res *APIGatewayProxyResponse, params url.Values) error {
		return nil
	}
	testRouter := NewRouter(testFallThroughHandler)
	testParams := url.Values{}

	Convey("addNode", t, func() {
		testNode := node{}
		Convey("Should add a node to the tree", func() {
			testNode.addNode("GET", "/path", testHandler)
			So(testNode.children, ShouldHaveLength, 1)
			So(testNode.component, ShouldBeEmpty)
		})

		Convey("Should add a node with a named path to the tree", func() {
			testRouter.Handle("GET", "/path/:named", testNamedHandler)
			node, _ := testRouter.tree.traverse(strings.Split("/path/foo", "/")[1:], testParams)
			So(node.isNamedParam, ShouldBeTrue)
			So(node.methods, ShouldHaveLength, 1)
		})

		Convey("Should be able to update a node", func() {
			testRouter.Handle("POST", "/path/:named", testNamedHandler)
			node, _ := testRouter.tree.traverse(strings.Split("/path/foo", "/")[1:], testParams)
			So(node.methods, ShouldHaveLength, 2)
			So(node.component, ShouldEqual, ":named")
		})
	})
}
