package framework

import (
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

	// TODO: Figure out how to run tests with XRay
	// testTaskerVal := 0
	// testCtx := context.Background()
	// testHandler := func(testCtx context.Context, evt *map[string]interface{}) error {
	// 	testTaskerVal = 1
	// 	return nil
	// }
	// testEvt := map[string]interface{}{
	// 	"_taskName": "test task",
	// 	"foo":       "bar",
	// }

	// os.Setenv("AWS_XRAY_CONTEXT_MISSING", "LOG_ERROR")
	// testTasker.Handle("test task", testHandler)
	// err := testTasker.LambdaHandler(testCtx, testEvt)

	// Convey("LambdaHandler", t, func() {
	// 	Convey("Should handle proper event", func() {
	// 		So(err, ShouldBeNil)
	// 		So(testTaskerVal, ShouldEqual, 1)
	// 	})
	// })
}
