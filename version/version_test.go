package version

import (
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	Convey("Should have a semantic version number", t, func() {
		So(Semantic, ShouldNotBeEmpty)
		pieces := strings.Split(Semantic, ".")
		So(pieces, ShouldHaveLength, 3)
	})

	Convey("Current should return the current semantic version number", t, func() {
		So(Current(), ShouldNotBeEmpty)
	})
}
