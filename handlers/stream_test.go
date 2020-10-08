package handlers

import (
	"github.com/companieshouse/chs-go-avro-schemas/data"
	"github.com/companieshouse/chs-streaming-api-frontend/unittesting"
	"net/http/httptest"

	"errors"
	"github.com/golang/mock/gomock"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var testFilingStream Stream

var (
	mockCtlr   *gomock.Controller
	req        *http.Request
	testStream Streaming
)

var StreamProcessHTTPTests = unittesting.Tests{List: unittesting.TestsList{
	"TestHeartbeatTimeoutAndRequestTimeoutHandledOK": unittesting.UnitTest{
		Title: "Trap Heartbeat timeout",
		Given: "Heartbeat timeout sent",
		Then:  "the heartbeat timeout is called",
		Vars:  unittesting.GenericMap{},
	}},
}

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (c *closeNotifyingRecorder) close() {
	c.closed <- true
}
func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}
func (c *closeNotifyingRecorder) Flush() {
	c.Flushed = true
}
func newCloseNotifyingRecorder() *closeNotifyingRecorder {
	return &closeNotifyingRecorder{httptest.NewRecorder(), make(chan bool, 1)}
}

// Test Stubs
func stubCheckIfDataNotPresentReturnError(sch data.ResourceChangedData) error {
	return errors.New("no data is available for the stream")
}

var requestTimeoutCalled bool

func stubHandleRequestTimeOut(contextID string) {
	requestTimeoutCalled = true
}

var processMessageCalled bool

func stubProcessMessage(contextID string, w http.ResponseWriter, stream Stream) (err error) {
	processMessageCalled = true
	return nil
}

func stubProcessMessageReturnError(contextID string, w http.ResponseWriter, stream Stream) (err error) {
	processMessageCalled = true
	return errors.New("error consuming message")
}

var heartbeatTimeoutCalled bool

func stubHandleHeartbeatTimeout(contextID string) {
	heartbeatTimeoutCalled = true
}

var processMessageFailedCalled bool

func stubProcessMessageFailed(contextID string, err error) {
	processMessageFailedCalled = true
}

var clientDisconnectCalled bool

func stubHandleClientDisconnect(contextID string) {
	clientDisconnectCalled = true
}

func initStreamProcessHTTPTest(t *testing.T) *unittesting.UnitTest {
	testID := unittesting.GetCallerName()
	unittesting.ResetInternalStatus()
	unitTest := StreamProcessHTTPTests.InitRunningTest(testID)

	// Initialise
	callCheckIfDataNotPresent = stubCheckIfDataNotPresentReturnError
	callRequestTimeOut = stubHandleRequestTimeOut
	callHeartbeatTimeout = stubHandleHeartbeatTimeout

	mockCtlr = gomock.NewController(t)

	callProcessMessageFailed = stubProcessMessageFailed
	callHandleClientDisconnect = stubHandleClientDisconnect

	var reqerr error
	req, reqerr = http.NewRequest("GET", "/filings", nil)
	if reqerr != nil {
		t.Errorf("Fail to create http request : %s", reqerr.Error())
	}

	requestTimeoutCalled = false
	processMessageCalled = false
	processMessageFailedCalled = false
	clientDisconnectCalled = false
	heartbeatTimeoutCalled = false

	//time in seconds
	testStream = Streaming{RequestTimeout: 100, HeartbeatInterval: 10}

	return &unitTest
}

//Test for Heart beat
func TestHeartbeatTimeoutAndRequestTimeoutHandledOK(t *testing.T) {

	unitTest := initStreamProcessHTTPTest(t)
	Convey(unitTest.GetGiven(), t, func() {
		Convey(unitTest.GetWhen(), func() {
			Convey(unitTest.GetThen(), func() {

				w := newCloseNotifyingRecorder()
				defer mockCtlr.Finish()

				//For test reset streaming request timeout and heartbeatInterval
				testStream.RequestTimeout = 3    //in seconds
				testStream.HeartbeatInterval = 1 //in seconds

				testStream.ProcessHTTP(w, req, testFilingStream)

				So(w.Code, ShouldEqual, 200)
				So(requestTimeoutCalled, ShouldEqual, true)
				So(processMessageCalled, ShouldEqual, false)
				So(clientDisconnectCalled, ShouldEqual, false)
				So(heartbeatTimeoutCalled, ShouldEqual, true)
			})
		})
	})
}
