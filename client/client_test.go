package client

import (
	"github.com/companieshouse/chs-streaming-api-frontend/broker"
	"github.com/companieshouse/chs.go/log"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"net/http"
	"strings"
	"sync"
	"testing"
)

type mockBroker struct {
	mock.Mock
}

func (b *mockBroker) Publish(msg string) {
	b.Called(msg)
}

type mockHttpClient struct {
	mock.Mock
}

func (c *mockHttpClient) Get(url string) (resp *http.Response, err error) {
	args := c.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

type mockBody struct {
	*strings.Reader
}

func (b *mockBody) Close() error {
	return nil
}

type mockLogger struct {
	mock.Mock
}

func (l *mockLogger) Error(err error, data ...log.Data) {
	l.Called(mock.Anything)
}

func (l *mockLogger) Info(msg string, data ...log.Data) {
	panic("implement me")
}

func (mockLogger) InfoR(req *http.Request, message string, data ...log.Data) {
	panic("implement me")
}

func TestNewClient(t *testing.T) {
	Convey("when a new client instance is created", t, func() {

		actual := NewClient("baseurl", "/path", &broker.CacheBroker{}, &http.Client{}, &mockLogger{})

		Convey("then a new client should be created", func() {
			So(actual, ShouldNotBeNil)
			So(actual.baseurl, ShouldEqual, "baseurl")
			So(actual.path, ShouldEqual, "/path")
			So(actual.broker, ShouldResemble, &broker.CacheBroker{})
			So(actual.httpClient, ShouldResemble, &http.Client{})
		})
	})
}

func TestPublishToBroker(t *testing.T) {
	Convey("given a mock broker and http client is called", t, func() {
		publisher := &mockBroker{}
		publisher.On("Publish", mock.Anything).Return()

		getter := &mockHttpClient{}
		getter.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200,
			Body: &mockBody{strings.NewReader("Test Data \n")},
		}, nil)

		logger := &mockLogger{}
		logger.On("Error", mock.Anything).Return(nil)

		client := NewClient("baseurl", "/path", publisher, getter, logger)
		client.wg = new(sync.WaitGroup)

		Convey("when a new message is published from cache broker", func() {
			client.wg.Add(1)
			client.Connect()
			client.wg.Wait()

			Convey("Then the message should be forwarded to the broker", func() {
				So(publisher.AssertCalled(t, "Publish", "Test Data \n"), ShouldBeTrue)
			})
		})
	})
}
