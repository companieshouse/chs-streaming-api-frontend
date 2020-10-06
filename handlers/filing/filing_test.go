package filing

import (
	"net/http"
	"testing"

	"github.com/companieshouse/chs-streaming-api/testdata"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitFiling(t *testing.T) {

	testStreaming := &Streaming{}

	Convey("Test successful transform of filing history", t, func() {
		req, err := http.NewRequest("GET", "/filings", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Accept", "*/*")
		testJSON := []byte(testdata.FilingJSON)
		err = testStreaming.Transform(req, &testJSON)
		So(err, ShouldBeNil)
	})

	Convey("Test Filter does not modify data in any way", t, func() {
		req, err := http.NewRequest("GET", "/filings", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Accept", "*/*")
		testJSONString := testdata.FilingJSON
		testJSON := []byte(testdata.FilingJSON)
		err = testStreaming.Filter(req, &testJSON)
		So(err, ShouldBeNil)
		So(string(testJSON), ShouldEqual, testJSONString)
	})
}
