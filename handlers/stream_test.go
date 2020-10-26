package handlers

import (
	"github.com/companieshouse/chs-streaming-api-frontend/logger"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"math"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

type mockContext struct {
	mock.Mock
}

type mockTimerFactory struct {
	C chan time.Time
	mock.Mock
}

// Mock Broker
type mockBroker struct {
	mock.Mock
}

func TestAMessagePublishedByTheBrokerIsWrittenToResponse(t *testing.T) {
	Convey("Given a broker is available", t, func() {
		subscription := make(chan string)
		cacheBrokerMock := &mockBroker{}
		cacheBrokerMock.On("Subscribe").Return(subscription, nil)
		cacheBrokerMock.On("Unsubscribe", subscription).Return(nil)
		timerFactory := &mockTimerFactory{}
		timerFactory.On("GetTimer", time.Duration(3)).Return(time.NewTimer(math.MaxInt64))
		timerFactory.On("GetTimer", time.Duration(1)).Return(time.NewTimer(math.MaxInt64))
		req := httptest.NewRequest("GET", "/filings", nil)

		w := newCloseNotifyingRecorder()

		//For test reset streaming request timeout and heartbeatInterval
		testStream := &Streaming{
			RequestTimeout:    3,
			HeartbeatInterval: 1,
			timerFactory:      timerFactory,
			Logger:            logger.NewLogger(),
			wg:                new(sync.WaitGroup),
		}
		testStream.wg.Add(1)

		go testStream.ProcessHTTP(w, req, cacheBrokerMock)

		Convey("When a message is published by the broker", func() {

			subscription <- "hello"
			testStream.wg.Wait()

			Convey("Then the message should be pushed to the user", func() {

				So(w.Code, ShouldEqual, 200)
				So(timerFactory.AssertCalled(t, "GetTimer", time.Duration(3)), ShouldBeTrue)
				So(timerFactory.AssertCalled(t, "GetTimer", time.Duration(1)), ShouldBeTrue)
				So(w.Body.Bytes(), ShouldResemble, []byte("hello"))
				So(cacheBrokerMock.AssertCalled(t, "Subscribe"), ShouldBeTrue)
			})
		})
	})
}

func TestUnsubscribeFromBrokerWhenUserDisconnects(t *testing.T) {
	Convey("Given a user has connected to the stream", t, func() {
		subscription := make(chan string)
		connectionClosed := make(chan struct{})
		context := &mockContext{}
		context.On("Done").Return(connectionClosed)
		cacheBrokerMock := &mockBroker{}
		cacheBrokerMock.On("Subscribe").Return(subscription, nil)
		cacheBrokerMock.On("Unsubscribe", subscription).Return(nil)
		timerFactory := &mockTimerFactory{}
		timerFactory.On("GetTimer", time.Duration(3)).Return(time.NewTimer(math.MaxInt64))
		timerFactory.On("GetTimer", time.Duration(1)).Return(time.NewTimer(math.MaxInt64))
		req := httptest.NewRequest("GET", "/filings", nil).WithContext(context)

		w := newCloseNotifyingRecorder()

		//For test reset streaming request timeout and heartbeatInterval
		testStream := &Streaming{
			RequestTimeout:    3,
			HeartbeatInterval: 1,
			timerFactory:      timerFactory,
			Logger:            logger.NewLogger(),
			wg:                new(sync.WaitGroup),
		}
		testStream.wg.Add(1)

		go testStream.ProcessHTTP(w, req, cacheBrokerMock)

		Convey("When the user disconnects from the stream", func() {
			connectionClosed <- struct{}{}
			testStream.wg.Wait()
			Convey("Then the user should be unsubscribed from the broker", func() {
				So(w.Code, ShouldEqual, 200)
				So(timerFactory.AssertCalled(t, "GetTimer", time.Duration(3)), ShouldBeTrue)
				So(timerFactory.AssertCalled(t, "GetTimer", time.Duration(1)), ShouldBeTrue)
				So(cacheBrokerMock.AssertCalled(t, "Subscribe"), ShouldBeTrue)
				So(cacheBrokerMock.AssertCalled(t, "Unsubscribe", subscription), ShouldBeTrue)
			})
		})
	})
}

func TestUnsubscribeFromBrokerIfConnectionExpired(t *testing.T) {
	Convey("Given a user has connected to the stream", t, func() {
		subscription := make(chan string)

		cacheBrokerMock := &mockBroker{}
		cacheBrokerMock.On("Subscribe").Return(subscription, nil)
		cacheBrokerMock.On("Unsubscribe", subscription).Return(nil)

		timerFactory := &mockTimerFactory{}
		timerFactory.On("GetTimer", time.Duration(3)).Return(time.NewTimer(0))
		timerFactory.On("GetTimer", time.Duration(1)).Return(time.NewTimer(math.MaxInt64))

		req := httptest.NewRequest("GET", "/filings", nil)

		w := newCloseNotifyingRecorder()

		//For test reset streaming request timeout and heartbeatInterval
		testStream := &Streaming{
			RequestTimeout:    3,
			HeartbeatInterval: 1,
			timerFactory:      timerFactory,
			Logger:            logger.NewLogger(),
			wg:                new(sync.WaitGroup),
		}
		testStream.wg.Add(1)

		go testStream.ProcessHTTP(w, req, cacheBrokerMock)
		Convey("When the connection times out", func() {
			testStream.wg.Wait()
			Convey("Then the user should be unsubscribed from the broker", func() {
				So(w.Code, ShouldEqual, 200)
				So(timerFactory.AssertCalled(t, "GetTimer", time.Duration(3)), ShouldBeTrue)
				So(timerFactory.AssertCalled(t, "GetTimer", time.Duration(1)), ShouldBeTrue)
				So(cacheBrokerMock.AssertCalled(t, "Subscribe"), ShouldBeTrue)
				So(cacheBrokerMock.AssertCalled(t, "Unsubscribe", subscription), ShouldBeTrue)
			})
		})
	})
}

func TestSendNewlineIfHeartbeat(t *testing.T) {
	Convey("Given a user has connected to the stream", t, func() {
		subscription := make(chan string)

		cacheBrokerMock := &mockBroker{}
		cacheBrokerMock.On("Subscribe").Return(subscription, nil)
		cacheBrokerMock.On("Unsubscribe", subscription).Return(nil)

		timerFactory := &mockTimerFactory{}
		timerFactory.On("GetTimer", time.Duration(3)).Return(time.NewTimer(math.MaxInt64))
		timerFactory.On("GetTimer", time.Duration(1)).Return(time.NewTimer(0))

		req := httptest.NewRequest("GET", "/filings", nil)

		w := newCloseNotifyingRecorder()

		//For test reset streaming request timeout and heartbeatInterval
		testStream := &Streaming{
			RequestTimeout:    3,
			HeartbeatInterval: 1,
			timerFactory:      timerFactory,
			Logger:            logger.NewLogger(),
			wg:                new(sync.WaitGroup),
		}
		testStream.wg.Add(1)

		go testStream.ProcessHTTP(w, req, cacheBrokerMock)
		Convey("When the connection times out", func() {
			testStream.wg.Wait()
			Convey("Then the user should be unsubscribed from the broker", func() {
				So(w.Code, ShouldEqual, 200)
				So(timerFactory.AssertCalled(t, "GetTimer", time.Duration(3)), ShouldBeTrue)
				So(timerFactory.AssertCalled(t, "GetTimer", time.Duration(1)), ShouldBeTrue)
				So(cacheBrokerMock.AssertCalled(t, "Subscribe"), ShouldBeTrue)
				So(w.Body.Bytes(), ShouldResemble, []byte("\n"))
			})
		})
	})
}

func newCloseNotifyingRecorder() *closeNotifyingRecorder {
	return &closeNotifyingRecorder{httptest.NewRecorder(), make(chan bool, 1)}
}

func (t *mockTimerFactory) GetTimer(duration time.Duration) *time.Timer {
	return t.Called(duration).Get(0).(*time.Timer)
}

func (c *mockContext) Done() <-chan struct{} {
	return c.Called().Get(0).(chan struct{})
}

func (c *mockContext) Value(v interface{}) interface{} {
	return c.Called(v).Get(0)
}

func (c *mockContext) Err() error {
	return c.Called().Error(0)
}

func (c *mockContext) Deadline() (time.Time, bool) {
	args := c.Called()
	return args.Get(0).(time.Time), args.Bool(1)
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

func (b *mockBroker) Subscribe() (chan string, error) {
	args := b.Called()
	return args.Get(0).(chan string), args.Error(1)
}

func (b *mockBroker) Unsubscribe(subscription chan string) error {
	args := b.Called(subscription)
	return args.Error(0)
}

func (b *mockBroker) Publish(msg string) {
	b.Called(msg)
}