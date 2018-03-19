package framework

import (
	"context"
	"encoding/json"
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
	testHandler := func(testCtx context.Context, evt *CloudWatchEvent) {
		testTaskerVal = 1
	}
	testEvt := CloudWatchEvent{
		DetailType: "Scheduled Event",
		Source:     "aws.events",
		Detail:     json.RawMessage("{}"),
		Resources:  []string{"arn:aws:events:us-east-1:123456789012:rule/MyScheduledRule"},
	}

	testTasker.Handle("MyScheduledRule", testHandler)
	testTasker.LambdaHandler(testCtx, testEvt)

	Convey("LambdaHandler", t, func() {
		Convey("Should handle proper event", func() {
			So(testTaskerVal, ShouldEqual, 1)
		})
	})
}
