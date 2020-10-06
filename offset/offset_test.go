package offset

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var offsetManager = NewOffset()

func TestUnitOffset(t *testing.T) {
	Convey("Test Parse returns 0 offset when empty string passed in", t, func() {
		offset, err := offsetManager.Parse("")
		So(err, ShouldBeNil)
		So(offset, ShouldEqual, 0)
	})

	Convey("Test Parse returns conversion error if non numberic string passed in", t, func() {
		offset, err := offsetManager.Parse("Fail")
		So(err, ShouldNotBeNil)
		So(offset, ShouldEqual, 0)
	})

	Convey("Test Parse returns err negative offset if negative number passed in", t, func() {
		offset, err := offsetManager.Parse("-5")
		So(err, ShouldEqual, ErrNegativeOffset)
		So(offset, ShouldEqual, 0)
	})

	Convey("Test Parse returns int64 if valid number pased in", t, func() {
		offset, err := offsetManager.Parse("5")
		So(err, ShouldBeNil)
		So(offset, ShouldEqual, 5)
	})
}
