package cmd

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestInitCmd(t *testing.T) {
	Convey("Should copy a boilerplate yaml config file", t, func() {
		testConfigFilePath := "./aegis-config.yaml.test"
		err := copyConfig(testConfigFilePath)
		So(err, ShouldBeNil)
		// file should exist
		_, err = os.Stat(testConfigFilePath)
		So(err, ShouldBeNil)

		Convey("Should not overwrite an existing file by the same name", func() {
			err = copyConfig(testConfigFilePath)
			So(err, ShouldNotBeNil)
		})

		// cleanup
		_ = os.Remove(testConfigFilePath)
	})

	Convey("Should copy a boilerplate .go source file", t, func() {
		testSrcFilePath := "./aegis-main.go.test"
		err := copySrc(testSrcFilePath)
		So(err, ShouldBeNil)
		// file should exist
		_, err = os.Stat(testSrcFilePath)
		So(err, ShouldBeNil)

		Convey("Should not overwrite an existing file by the same name", func() {
			err = copySrc(testSrcFilePath)
			So(err, ShouldNotBeNil)
		})

		// cleanup
		_ = os.Remove(testSrcFilePath)
	})
}
