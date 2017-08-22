package cmd

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRootCmd(t *testing.T) {
	Convey("Should initialize default config options", t, func() {
		InitConfig()
		So(cfg.AWS.Region, ShouldEqual, "us-east-1")
	})
}
