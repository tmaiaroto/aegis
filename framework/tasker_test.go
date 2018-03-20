package framework

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTasker(t *testing.T) {

	testTasker := NewTasker()
	Convey("NewTasker", t, func() {
		Convey("Should create a new Tasker", func() {
			So(testTasker, ShouldNotBeNil)
		})
	})

	testTaskerVal := 0
	testCtx := context.Background()
	testHandler := func(testCtx context.Context, evt *map[string]interface{}) error {
		testTaskerVal = 1
		return nil
	}
	testEvt := map[string]interface{}{
		"_taskName": "test task",
		"foo":       "bar",
	}

	testTasker.Handle("test task", testHandler)
	testTasker.LambdaHandler(testCtx, testEvt)

	Convey("LambdaHandler", t, func() {
		Convey("Should handle proper event", func() {
			So(testTaskerVal, ShouldEqual, 1)
		})
	})
}
