package handlers

import (
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

var (
	mockCtlr   *gomock.Controller
	req        *http.Request
	testStream Streaming
)

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
var requestTimeoutCalled bool

func stubHandleRequestTimeOut(contextID string) {
	requestTimeoutCalled = true
}

var heartbeatTimeoutCalled bool

func stubHandleHeartbeatTimeout(contextID string) {
	heartbeatTimeoutCalled = true
}

var clientDisconnectCalled bool

func stubHandleClientDisconnect(contextID string) {
	clientDisconnectCalled = true
}

type mockBroker struct {
	mock.Mock
}

func (b *mockBroker) Subscribe() (chan string, error) {
	subscription := make(chan string)
	return subscription, nil
}

func (b *mockBroker) Unsubscribe(chan string) error {
	return nil
}

func (b *mockBroker) Publish(msg string) {
	b.Called(msg)
}

func initStreamProcessHTTPTest(t *testing.T) {

	// Initialise
	callRequestTimeOut = stubHandleRequestTimeOut
	callHeartbeatTimeout = stubHandleHeartbeatTimeout
	callHandleClientDisconnect = stubHandleClientDisconnect

	mockCtlr = gomock.NewController(t)

	var reqerr error
	req, reqerr = http.NewRequest("GET", "/filings", nil)
	if reqerr != nil {
		t.Errorf("Fail to create http request : %s", reqerr.Error())
	}

	requestTimeoutCalled = false
	clientDisconnectCalled = false
	heartbeatTimeoutCalled = false

	//time in seconds
	testStream = Streaming{RequestTimeout: 100, HeartbeatInterval: 10}
}

func TestHeartbeatTimeoutAndRequestTimeoutHandledOK(t *testing.T) {

	initStreamProcessHTTPTest(t)

	Convey("given a broker is available", t, func() {

		subscription := make(chan string)
		cacheBrokerMock := &mockBroker{}
		cacheBrokerMock.On("Subscribe").Return(subscription, nil)
		cacheBrokerMock.On("Unsubscribe", subscription).Return(nil)

		Convey("when a stream request is processed", func() {
			w := newCloseNotifyingRecorder()
			defer mockCtlr.Finish()

			//For test reset streaming request timeout and heartbeatInterval
			testStream.RequestTimeout = 3    //in seconds
			testStream.HeartbeatInterval = 1 //in seconds
			testStream.Broker = cacheBrokerMock
			testStream.wg = new(sync.WaitGroup)

			testStream.ProcessHTTP(w, req)

			Convey("then broker should be invoked", func() {

				So(w.Code, ShouldEqual, 200)
				So(requestTimeoutCalled, ShouldEqual, true)
				So(clientDisconnectCalled, ShouldEqual, false)
				So(heartbeatTimeoutCalled, ShouldEqual, true)
			})
		})
	})
}
