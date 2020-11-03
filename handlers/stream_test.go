package handlers

import (
	"github.com/companieshouse/chs-streaming-api-frontend/client"
	"github.com/companieshouse/chs-streaming-api-frontend/factory"
	"github.com/companieshouse/chs-streaming-api-frontend/logger"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
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

type mockClientFactory struct {
	mock.Mock
}

type mockClient struct {
	mock.Mock
}

type mockPublisher struct {
	mock.Mock
}

type mockBroker struct {
	mock.Mock
}

type mockTimerFactory struct {
	mock.Mock
}

type mockTimer struct {
	mock.Mock
	events chan bool
}

func TestAMessagePublishedByTheBrokerIsWrittenToResponse(t *testing.T) {
	Convey("Given a broker is available", t, func() {
		subscription := make(chan string)
		cacheBrokerMock := &mockBroker{}
		cacheBrokerMock.On("Subscribe").Return(subscription, nil)
		cacheBrokerMock.On("Unsubscribe", subscription).Return(nil)
		req := httptest.NewRequest("GET", "/filings", nil)

		requestTimer := &mockTimer{}
		requestTimer.On("Start").Return()
		requestTimer.On("Elapsed").Return()

		heartBeatTimer := &mockTimer{}
		heartBeatTimer.On("Start").Return()
		heartBeatTimer.On("Elapsed").Return()

		timerFactory := &mockTimerFactory{}
		timerFactory.On("GetTimer", mock.Anything).Return(requestTimer).Once()
		timerFactory.On("GetTimer", mock.Anything).Return(heartBeatTimer).Once()

		w := newCloseNotifyingRecorder()

		//For test reset streaming request timeout and heartbeatInterval
		testStream := &Streaming{
			RequestTimeout:    10,
			HeartbeatInterval: 8,
			Logger:            logger.NewLogger(),
			TimerFactory:      timerFactory,
			wg:                new(sync.WaitGroup),
		}
		testStream.wg.Add(1)

		go testStream.ProcessHTTP(w, req, "", cacheBrokerMock)

		Convey("When a message is published by the broker", func() {

			subscription <- "hello"
			testStream.wg.Wait()

			Convey("Then the message should be pushed to the user", func() {

				So(w.Code, ShouldEqual, 200)
				So(w.Body.Bytes(), ShouldResemble, []byte("hello"))
				So(cacheBrokerMock.AssertCalled(t, "Subscribe"), ShouldBeTrue)
				So(timerFactory.AssertCalled(t, "GetTimer", 10*time.Second), ShouldBeTrue)
				So(timerFactory.AssertCalled(t, "GetTimer", 8*time.Second), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
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
		req := httptest.NewRequest("GET", "/filings", nil).WithContext(context)

		requestTimer := &mockTimer{}
		requestTimer.On("Start").Return()
		requestTimer.On("Elapsed").Return(make(chan bool))
		requestTimer.On("Stop").Return()

		heartBeatTimer := &mockTimer{}
		heartBeatTimer.On("Start").Return()
		heartBeatTimer.On("Elapsed").Return(make(chan bool))
		heartBeatTimer.On("Stop").Return()

		timerFactory := &mockTimerFactory{}
		timerFactory.On("GetTimer", mock.Anything).Return(requestTimer).Once()
		timerFactory.On("GetTimer", mock.Anything).Return(heartBeatTimer).Once()

		w := newCloseNotifyingRecorder()

		//For test reset streaming request timeout and heartbeatInterval
		testStream := &Streaming{
			RequestTimeout:    10,
			HeartbeatInterval: 10,
			TimerFactory:      timerFactory,
			Logger:            logger.NewLogger(),
			wg:                new(sync.WaitGroup),
		}
		testStream.wg.Add(1)

		go testStream.ProcessHTTP(w, req, "", cacheBrokerMock)

		Convey("When the user disconnects from the stream", func() {
			connectionClosed <- struct{}{}
			testStream.wg.Wait()
			Convey("Then the user should be unsubscribed from the broker", func() {
				So(w.Code, ShouldEqual, 200)
				So(cacheBrokerMock.AssertCalled(t, "Subscribe"), ShouldBeTrue)
				So(cacheBrokerMock.AssertCalled(t, "Unsubscribe", subscription), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Stop"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Stop"), ShouldBeTrue)
			})
		})
	})
}

func TestUnsubscribeFromBrokerIfConnectionExpired(t *testing.T) {
	Convey("Given a user has connected to the stream", t, func() {
		notifyTimeout := make(chan bool)
		subscription := make(chan string)

		cacheBrokerMock := &mockBroker{}
		cacheBrokerMock.On("Subscribe").Return(subscription, nil)
		cacheBrokerMock.On("Unsubscribe", subscription).Return(nil)

		req := httptest.NewRequest("GET", "/filings", nil)

		requestTimer := &mockTimer{events: notifyTimeout}
		requestTimer.On("Start").Return()
		requestTimer.On("Elapsed").Return(notifyTimeout)
		requestTimer.On("Stop").Return()

		heartBeatTimer := &mockTimer{}
		heartBeatTimer.On("Start").Return()
		heartBeatTimer.On("Elapsed").Return()
		heartBeatTimer.On("Stop").Return()

		timerFactory := &mockTimerFactory{}
		timerFactory.On("GetTimer", mock.Anything).Return(requestTimer).Once()
		timerFactory.On("GetTimer", mock.Anything).Return(heartBeatTimer).Once()

		w := newCloseNotifyingRecorder()

		//For test reset streaming request timeout and heartbeatInterval
		testStream := &Streaming{
			RequestTimeout:    3,
			HeartbeatInterval: 10,
			TimerFactory:      timerFactory,
			Logger:            logger.NewLogger(),
			wg:                new(sync.WaitGroup),
		}
		testStream.wg.Add(1)

		go testStream.ProcessHTTP(w, req, "", cacheBrokerMock)
		Convey("When the connection times out", func() {
			notifyTimeout <- true
			testStream.wg.Wait()
			Convey("Then the user should be unsubscribed from the broker", func() {
				So(w.Code, ShouldEqual, 200)
				So(cacheBrokerMock.AssertCalled(t, "Subscribe"), ShouldBeTrue)
				So(cacheBrokerMock.AssertCalled(t, "Unsubscribe", subscription), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Stop"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Stop"), ShouldBeTrue)
			})
		})
	})
}

func TestSendNewlineIfHeartbeat(t *testing.T) {
	Convey("Given a user has connected to the stream", t, func() {
		notifyHeartbeat := make(chan bool)
		subscription := make(chan string)

		cacheBrokerMock := &mockBroker{}
		cacheBrokerMock.On("Subscribe").Return(subscription, nil)
		cacheBrokerMock.On("Unsubscribe", subscription).Return(nil)

		req := httptest.NewRequest("GET", "/filings", nil)

		requestTimer := &mockTimer{}
		requestTimer.On("Start").Return()
		requestTimer.On("Elapsed").Return()
		requestTimer.On("Stop").Return()

		heartBeatTimer := &mockTimer{events: notifyHeartbeat}
		heartBeatTimer.On("Start").Return()
		heartBeatTimer.On("Elapsed").Return(notifyHeartbeat)
		heartBeatTimer.On("Stop").Return()
		heartBeatTimer.On("Reset").Return()

		timerFactory := &mockTimerFactory{}
		timerFactory.On("GetTimer", mock.Anything).Return(requestTimer).Once()
		timerFactory.On("GetTimer", mock.Anything).Return(heartBeatTimer).Once()

		w := newCloseNotifyingRecorder()

		//For test reset streaming request timeout and heartbeatInterval
		testStream := &Streaming{
			RequestTimeout:    10,
			HeartbeatInterval: 3,
			TimerFactory:      timerFactory,
			Logger:            logger.NewLogger(),
			wg:                new(sync.WaitGroup),
		}
		testStream.wg.Add(1)

		go testStream.ProcessHTTP(w, req, "", cacheBrokerMock)
		Convey("When the heartbeat timer elapses", func() {
			notifyHeartbeat <- true
			testStream.wg.Wait()
			Convey("Then a newline should be sent to the connected user", func() {
				So(w.Code, ShouldEqual, 200)
				So(cacheBrokerMock.AssertCalled(t, "Subscribe"), ShouldBeTrue)
				So(w.Body.Bytes(), ShouldResemble, []byte("\n"))
				So(requestTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Reset"), ShouldBeTrue)
			})
		})
	})
}

func TestOffsetSpecified(t *testing.T) {
	Convey("Given a user has connected to the stream", t, func() {
		subscription := make(chan string)

		publisher := &mockPublisher{}
		publisher.On("Publish", mock.Anything).Return()
		publisher.On("Subscribe").Return(subscription, nil)
		publisher.On("Unsubscribe", mock.Anything).Return(nil)

		serviceClient := &mockClient{}
		serviceClient.On("Connect").Return(&client.ResponseStatus{Code: 200})
		serviceClient.On("SetOffset", mock.Anything).Return()

		clientFactory := &mockClientFactory{}
		clientFactory.On("GetClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(serviceClient)

		requestTimer := &mockTimer{}
		requestTimer.On("Start").Return()
		requestTimer.On("Elapsed").Return()

		heartBeatTimer := &mockTimer{}
		heartBeatTimer.On("Start").Return()
		heartBeatTimer.On("Elapsed").Return()

		timerFactory := &mockTimerFactory{}
		timerFactory.On("GetTimer", mock.Anything).Return(requestTimer).Once()
		timerFactory.On("GetTimer", mock.Anything).Return(heartBeatTimer).Once()

		//baseurl string, path string, publisher serviceClient.Publishable, logger logger.Logger
		req := httptest.NewRequest("GET", "/filings?timepoint=1", nil)

		w := newCloseNotifyingRecorder()

		//For test reset streaming request timeout and heartbeatInterval
		testStream := &Streaming{
			CacheBrokerURL:    "baseurl",
			RequestTimeout:    10,
			HeartbeatInterval: 10,
			ClientFactory:     clientFactory,
			TimerFactory:      timerFactory,
			Logger:            logger.NewLogger(),
			wg:                new(sync.WaitGroup),
		}
		testStream.wg.Add(1)

		go testStream.ProcessHTTP(w, req, "/filings", publisher)

		Convey("When offsets beyond a specific timepoint are consumed from the cache service", func() {
			subscription <- "hello"
			testStream.wg.Wait()
			Convey("Then the message should be pushed to the user", func() {
				So(w.Code, ShouldEqual, 200)
				So(w.Body.Bytes(), ShouldResemble, []byte("hello"))
				So(clientFactory.AssertCalled(t, "GetClient", "baseurl", "/filings", publisher, mock.Anything), ShouldBeTrue)
				So(serviceClient.AssertCalled(t, "SetOffset", "1"), ShouldBeTrue)
				So(serviceClient.AssertCalled(t, "Connect"), ShouldBeTrue)
				So(publisher.AssertCalled(t, "Subscribe"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
			})
		})
	})
}

func TestCloseClientWhenUserDisconnects(t *testing.T) {
	Convey("Given a user has connected to the stream", t, func() {
		subscription := make(chan string)
		connectionClosed := make(chan struct{})

		publisher := &mockPublisher{}
		publisher.On("Publish", mock.Anything).Return()
		publisher.On("Subscribe").Return(subscription, nil)
		publisher.On("Unsubscribe", mock.Anything).Return(nil)

		serviceClient := &mockClient{}
		serviceClient.On("Connect").Return(&client.ResponseStatus{Code: 200})
		serviceClient.On("SetOffset", mock.Anything).Return()
		serviceClient.On("Close").Return()

		clientFactory := &mockClientFactory{}
		clientFactory.On("GetClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(serviceClient)

		requestTimer := &mockTimer{}
		requestTimer.On("Start").Return()
		requestTimer.On("Elapsed").Return()
		requestTimer.On("Stop").Return()

		heartBeatTimer := &mockTimer{}
		heartBeatTimer.On("Start").Return()
		heartBeatTimer.On("Elapsed").Return()
		heartBeatTimer.On("Stop").Return()

		timerFactory := &mockTimerFactory{}
		timerFactory.On("GetTimer", mock.Anything).Return(requestTimer).Once()
		timerFactory.On("GetTimer", mock.Anything).Return(heartBeatTimer).Once()

		context := &mockContext{}
		context.On("Done").Return(connectionClosed)
		req := httptest.NewRequest("GET", "/filings?timepoint=1", nil).WithContext(context)

		w := newCloseNotifyingRecorder()

		//For test reset streaming request timeout and heartbeatInterval
		testStream := &Streaming{
			CacheBrokerURL:    "baseurl",
			RequestTimeout:    10,
			HeartbeatInterval: 10,
			ClientFactory:     clientFactory,
			TimerFactory:      timerFactory,
			Logger:            logger.NewLogger(),
			wg:                new(sync.WaitGroup),
		}
		testStream.wg.Add(1)

		go testStream.ProcessHTTP(w, req, "/filings", publisher)

		Convey("When the user disconnects from the stream", func() {
			connectionClosed <- struct{}{}
			testStream.wg.Wait()
			Convey("Then the user should be unsubscribed from the broker and the client should be closed", func() {
				So(w.Code, ShouldEqual, 200)
				So(clientFactory.AssertCalled(t, "GetClient", "baseurl", "/filings", publisher, mock.Anything), ShouldBeTrue)
				So(serviceClient.AssertCalled(t, "SetOffset", "1"), ShouldBeTrue)
				So(serviceClient.AssertCalled(t, "Connect"), ShouldBeTrue)
				So(publisher.AssertCalled(t, "Subscribe"), ShouldBeTrue)
				So(publisher.AssertCalled(t, "Unsubscribe", subscription), ShouldBeTrue)
				So(serviceClient.AssertCalled(t, "Close"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Stop"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Stop"), ShouldBeTrue)
			})
		})
	})
}

func TestCloseClientIfConnectionExpired(t *testing.T) {
	Convey("Given a user has connected to the stream", t, func() {
		subscription := make(chan string)
		connectionClosed := make(chan struct{})

		publisher := &mockPublisher{}
		publisher.On("Publish", mock.Anything).Return()
		publisher.On("Subscribe").Return(subscription, nil)
		publisher.On("Unsubscribe", mock.Anything).Return(nil)

		serviceClient := &mockClient{}
		serviceClient.On("Connect").Return(&client.ResponseStatus{Code: 200})
		serviceClient.On("SetOffset", mock.Anything).Return()
		serviceClient.On("Close").Return()

		clientFactory := &mockClientFactory{}
		clientFactory.On("GetClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(serviceClient)

		context := &mockContext{}
		context.On("Done").Return(connectionClosed)

		requestTimer := &mockTimer{}
		requestTimer.On("Start").Return()
		requestTimer.On("Elapsed").Return()
		requestTimer.On("Stop").Return()

		heartBeatTimer := &mockTimer{}
		heartBeatTimer.On("Start").Return()
		heartBeatTimer.On("Elapsed").Return()
		heartBeatTimer.On("Stop").Return()

		timerFactory := &mockTimerFactory{}
		timerFactory.On("GetTimer", mock.Anything).Return(requestTimer).Once()
		timerFactory.On("GetTimer", mock.Anything).Return(heartBeatTimer).Once()

		req := httptest.NewRequest("GET", "/filings?timepoint=1", nil).WithContext(context)

		w := newCloseNotifyingRecorder()

		//For test reset streaming request timeout and heartbeatInterval
		testStream := &Streaming{
			CacheBrokerURL:    "baseurl",
			RequestTimeout:    10,
			HeartbeatInterval: 10,
			ClientFactory:     clientFactory,
			TimerFactory:      timerFactory,
			Logger:            logger.NewLogger(),
			wg:                new(sync.WaitGroup),
		}
		testStream.wg.Add(1)

		go testStream.ProcessHTTP(w, req, "/filings", publisher)

		Convey("When the user disconnects from the stream", func() {
			connectionClosed <- struct{}{}
			testStream.wg.Wait()
			Convey("Then the user should be unsubscribed from the broker and the client should be closed", func() {
				So(w.Code, ShouldEqual, 200)
				So(clientFactory.AssertCalled(t, "GetClient", "baseurl", "/filings", publisher, mock.Anything), ShouldBeTrue)
				So(serviceClient.AssertCalled(t, "SetOffset", "1"), ShouldBeTrue)
				So(serviceClient.AssertCalled(t, "Connect"), ShouldBeTrue)
				So(publisher.AssertCalled(t, "Subscribe"), ShouldBeTrue)
				So(publisher.AssertCalled(t, "Unsubscribe", subscription), ShouldBeTrue)
				So(serviceClient.AssertCalled(t, "Close"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
				So(requestTimer.AssertCalled(t, "Stop"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Start"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Elapsed"), ShouldBeTrue)
				So(heartBeatTimer.AssertCalled(t, "Stop"), ShouldBeTrue)
			})
		})
	})
}

func newCloseNotifyingRecorder() *closeNotifyingRecorder {
	return &closeNotifyingRecorder{httptest.NewRecorder(), make(chan bool, 1)}
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

func (c *mockClientFactory) GetClient(baseurl string, path string, publisher client.Publishable, logger logger.Logger) factory.Connectable {
	return c.Called(baseurl, path, publisher, logger).Get(0).(factory.Connectable)
}

func (c *mockClient) Connect() *client.ResponseStatus {
	return c.Called().Get(0).(*client.ResponseStatus)
}

func (c *mockClient) SetOffset(offset string) {
	c.Called(offset)
}

func (c *mockClient) Close() {
	c.Called()
}

func (p *mockPublisher) Publish(msg string) {
	p.Called(msg)
}

func (p *mockPublisher) Subscribe() (chan string, error) {
	args := p.Called()
	return args.Get(0).(chan string), args.Error(1)
}

func (p *mockPublisher) Unsubscribe(subscription chan string) error {
	return p.Called(subscription).Error(0)
}

func (f *mockTimerFactory) GetTimer(duration time.Duration) factory.Elapseable {
	return f.Called(duration).Get(0).(factory.Elapseable)
}

func (t *mockTimer) Start() {
	t.Called()
}

func (t *mockTimer) Elapsed() <-chan bool {
	t.Called()
	return t.events
}

func (t *mockTimer) Reset() {
	t.Called()
}

func (t *mockTimer) Stop() {
	t.Called()
}
