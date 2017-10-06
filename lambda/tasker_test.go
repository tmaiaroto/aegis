package lambda

//. "github.com/smartystreets/goconvey/convey"

// func TestTasker(t *testing.T) {
// 	pr, pw := io.Pipe()
// 	TaskerIn = pw
// 	TaskerOut = pr

// 	testHandler := func(ctx *Context, evt *Event, params url.Values) {}
// 	// testParams := url.Values{}
// 	testTasker := NewTasker()
// 	Convey("NewTasker", t, func() {
// 		Convey("Should create a new Tasker", func() {
// 			So(testTasker, ShouldNotBeNil)
// 		})
// 	})

// 	testTasker.Handle("taskName", testHandler)

// 	testTasker.Listen()
// }
