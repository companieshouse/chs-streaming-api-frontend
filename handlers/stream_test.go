package handlers

import (
	"errors"
	"github.com/companieshouse/chs-streaming-api-frontend/offset"
	"github.com/companieshouse/chs-streaming-api-frontend/unittesting"
	. "github.com/smartystreets/goconvey/convey"
	"net/http/httptest"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/companieshouse/chs-go-avro-schemas/data"

	"github.com/companieshouse/chs-streaming-api-frontend/mocks/consumer"
	"github.com/companieshouse/chs-streaming-api-frontend/mocks/offset"

	"github.com/golang/mock/gomock"
	"net/http"
)

var testFilingStream Stream

const dummyTopic = "test-topic"

var dummyBroker = []string{"testBroker1"}

var (
	ChMessages = make(chan *sarama.ConsumerMessage, 3)

	mockCtlr            *gomock.Controller
	mockOffset          *mock_offset.MockInterface
	mockConsumerFactory *mock_consumer.MockFactoryInterface
	mockConsumer        *mock_consumer.MockInterface

	req        *http.Request
	testOffset int64
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

var StreamProcessHTTPTests = unittesting.Tests{List: unittesting.TestsList{
	"TestTimeOutHandledOK": unittesting.UnitTest{
		Title: "TimeOut Handled OK",
		Given: "when a timeout is handled the main loop is exited",
		Then:  "the timeout function is called",
		Vars:  unittesting.GenericMap{},
	},
	"TestHeartbeatHandledOK": unittesting.UnitTest{
		Title: "Trap Heartbeat timeout",
		Given: "Heartbeat timeout sent",
		Then:  "the heartbeat timeout is called",
		Vars:  unittesting.GenericMap{},
	},
	"TestFailWithInvalidOffset": unittesting.UnitTest{
		Title: "Test user supplied invalid offset",
		Given: "we receive an invalid offset from the user",
		Then:  "client receives a HTTP 416 error code",
		Vars:  unittesting.GenericMap{},
	},
	"TestFailToParseOffset": unittesting.UnitTest{
		Title: "Test fail to parse user supplied offset",
		Given: "we receive an offset from the user that we can not parse",
		Then:  "client receives a HTTP 400 error code",
		Vars:  unittesting.GenericMap{},
	},},
}

func initStreamProcessHTTPTest(t *testing.T) *unittesting.UnitTest {
	testID := unittesting.GetCallerName()
	unittesting.ResetInternalStatus()
	unitTest := StreamProcessHTTPTests.InitRunningTest(testID)

	// Initialise
	callCheckIfDataNotPresent = stubCheckIfDataNotPresentReturnError
	callLogHandleTimeOut = stubLogHandleTimeOut
	callProcessMessage = stubProcessMessage
	callProcessMessageFailed = stubProcessMessageFailed
	callLogHandleClientDisconnect = stubLogHandleClientDisconnect
	callLogHeartbeatTimeout = stubLogHeartbeatTimeout

	mockCtlr = gomock.NewController(t)
	mockOffset = mock_offset.NewMockInterface(mockCtlr)
	mockConsumerFactory = mock_consumer.NewMockFactoryInterface(mockCtlr)
	mockConsumer = mock_consumer.NewMockInterface(mockCtlr)

	var reqerr error
	req, reqerr = http.NewRequest("GET", "/filings", nil)
	if reqerr != nil {
		t.Errorf("Fail to create http request : %s", reqerr.Error())
	}

	testOffset = int64(1)
	timeoutCalled = false
	processMessageCalled = false
	processMessageFailedCalled = false
	clientDisconnectCalled = false
	heartbeatTimeoutCalled = false

	testStream = Streaming{BrokerAddr: dummyBroker, ConsumerFactory: mockConsumerFactory, Offset: mockOffset}

	return &unitTest
}

func stubCheckIfDataNotPresentReturnError(sch data.ResourceChangedData) error {
	return errors.New("no data is available for the stream")
}

var timeoutCalled bool

func stubLogHandleTimeOut(contextID string) {
	timeoutCalled = true
}

var processMessageCalled bool

func stubProcessMessage(contextID string, m *sarama.ConsumerMessage, w http.ResponseWriter, stream Stream) (err error) {
	processMessageCalled = true
	return nil
}

var processMessageFailedCalled bool

func stubProcessMessageFailed(contextID string, err error) {
	processMessageFailedCalled = true
}

var clientDisconnectCalled bool

func stubLogHandleClientDisconnect(contextID string) {
	clientDisconnectCalled = true
}

var heartbeatTimeoutCalled bool

func stubLogHeartbeatTimeout(contextID string) {
	heartbeatTimeoutCalled = true
}

// Request Timeout
func TestTimeOutHandledOK(t *testing.T) {

	unitTest := initStreamProcessHTTPTest(t)
	Convey(unitTest.GetGiven(), t, func() {
		Convey(unitTest.GetWhen(), func() {
			Convey(unitTest.GetThen(), func() {

				w := newCloseNotifyingRecorder()
				defer mockCtlr.Finish()

				testStream.RequestTimeout = 1 // seconds
				testStream.HeartbeatInterval = 5000

				mockOffset.EXPECT().Parse("").Return(testOffset, nil)
				mockOffset.EXPECT().IsValid(dummyBroker, dummyTopic, testOffset).Return(nil)
				mockConsumerFactory.EXPECT().Initialise(dummyTopic, dummyBroker).Return(mockConsumer)
				mockConsumer.EXPECT().ConsumePartition(int32(0), testOffset).Return(nil)
				mockConsumer.EXPECT().Messages().Return(ChMessages)

				testStream.ProcessHTTP(dummyTopic, w, req, testFilingStream)

				So(w.Code, ShouldEqual, 200)
				So(timeoutCalled, ShouldEqual, true)
				So(processMessageCalled, ShouldEqual, false)
				So(clientDisconnectCalled, ShouldEqual, false)
				So(heartbeatTimeoutCalled, ShouldEqual, false)
			})
		})
	})
}

//Test for Heart beat
func TestHeartbeatHandledOK(t *testing.T) {

	unitTest := initStreamProcessHTTPTest(t)
	Convey(unitTest.GetGiven(), t, func() {
		Convey(unitTest.GetWhen(), func() {
			Convey(unitTest.GetThen(), func() {

				w := newCloseNotifyingRecorder()
				defer mockCtlr.Finish()

				testStream.RequestTimeout = 3 // seconds
				testStream.HeartbeatInterval = 1

				mockOffset.EXPECT().Parse("").Return(testOffset, nil)
				mockOffset.EXPECT().IsValid(dummyBroker, dummyTopic, testOffset).Return(nil)
				mockConsumerFactory.EXPECT().Initialise(dummyTopic, dummyBroker).Return(mockConsumer)
				mockConsumer.EXPECT().ConsumePartition(int32(0), testOffset).Return(nil)
				mockConsumer.EXPECT().Messages().Return(ChMessages)

				testStream.ProcessHTTP(dummyTopic, w, req, testFilingStream)
				So(w.Code, ShouldEqual, 200)
				So(timeoutCalled, ShouldEqual, true)
				So(processMessageCalled, ShouldEqual, false)
				So(clientDisconnectCalled, ShouldEqual, false)
				So(heartbeatTimeoutCalled, ShouldEqual, true)
			})
		})
	})
}

// Invalid Offset - Out of range
func TestFailWithInvalidOffset(t *testing.T) {

	unitTest := initStreamProcessHTTPTest(t)
	Convey(unitTest.GetGiven(), t, func() {
		Convey(unitTest.GetWhen(), func() {
			Convey(unitTest.GetThen(), func() {

				w := newCloseNotifyingRecorder()
				defer mockCtlr.Finish()

				mockOffset.EXPECT().Parse("").Return(testOffset, nil)
				mockOffset.EXPECT().IsValid(dummyBroker, dummyTopic, testOffset).Return(offset.ErrOutOfRange)

				testStream.ProcessHTTP(dummyTopic, w, req, testFilingStream)

				So(w.Code, ShouldEqual, 416) //returns Request out of range
				So(w.ResponseRecorder.Body.String(), ShouldNotEqual, nil)
			})
		})
	})
}

// Invalid Offset - Bad request
func TestFailToParseOffset(t *testing.T) {

	unitTest := initStreamProcessHTTPTest(t)
	Convey(unitTest.GetGiven(), t, func() {
		Convey(unitTest.GetWhen(), func() {
			Convey(unitTest.GetThen(), func() {

				w := newCloseNotifyingRecorder()
				defer mockCtlr.Finish()

				mockOffset.EXPECT().Parse("").Return(testOffset, errors.New("fail to parse offset"))

				testStream.ProcessHTTP(dummyTopic, w, req, testFilingStream)

				So(w.Code, ShouldEqual, 400) //Bad Request
				So(w.ResponseRecorder.Body.String(), ShouldNotEqual, nil)
			})
		})
	})
}
