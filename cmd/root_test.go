package cmd

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestRootCmd(t *testing.T) {
	Convey("Should initialize default config options", t, func() {
		initConfig()
		So(cfg.AWS.Region, ShouldEqual, "us-east-1")
	})
}
