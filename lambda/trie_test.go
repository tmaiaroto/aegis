package lambda

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/url"
	"strings"
	"testing"
)

func TestTrie(t *testing.T) {
	testFallThroughHandler := func(ctx *Context, evt *Event, res *ProxyResponse, params url.Values) {}
	testHandler := func(ctx *Context, evt *Event, res *ProxyResponse, params url.Values) {}
	testNamedHandler := func(ctx *Context, evt *Event, res *ProxyResponse, params url.Values) {}
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
