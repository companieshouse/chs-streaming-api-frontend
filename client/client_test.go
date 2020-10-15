package client

import (
	"github.com/companieshouse/chs-streaming-api-frontend/broker"
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

func TestNewClient(t *testing.T) {
	Convey("given a new client instance is created", t, func() {
		actual := NewClient("baseurl", &broker.CacheBroker{}, &http.Client{})
		Convey("then a new client should be created", func() {
			So(actual, ShouldNotBeNil)
			So(actual.baseurl, ShouldEqual, "baseurl")
			So(actual.publisher, ShouldResemble, &broker.CacheBroker{})
			So(actual.client, ShouldResemble, &http.Client{})
		})
	})
}

func TestPublishToBroker(t *testing.T) {
	Convey("given a mock broker and http client is called", t, func() {
		publisher := &mockBroker{}
		publisher.On("Publish", mock.Anything).Return()

		getter := &mockHttpClient{}
		getter.On("Get", mock.Anything).Return(&http.Response{StatusCode: 200,
			Body: &mockBody{strings.NewReader("{\"data\":\"{\\\"greetings\\\":\\\"hello\\\"}\",\"offset\":43}\n")},
		}, nil)

		client := NewClient("baseurl", publisher, getter)
		client.Wg = new(sync.WaitGroup)

		Convey("when a new message is published from cache broker", func() {
			client.Wg.Add(1)
			client.Connect()
			client.Wg.Wait()
			Convey("Then the message should be forwarded to the broker", func() {
				So(publisher.AssertCalled(t, "Publish", "{\"greetings\":\"hello\"}"), ShouldBeTrue)
			})
		})
	})
}
