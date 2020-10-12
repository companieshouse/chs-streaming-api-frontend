package client

import (
	"github.com/companieshouse/chs-streaming-api-frontend/broker"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"net/http"
	"strings"
	"testing"
)

type mockBroker struct{
	mock.Mock
}

func (b *mockBroker) Publish(msg string) {
	b.Called(msg)
}

type mockHttpClient struct{
	mock.Mock
}

func (c *mockHttpClient) Get(url string) (resp *http.Response, err error) {
	args := c.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

type mockBody struct{
	*strings.Reader
}

func (b *mockBody) Close() error {
	return nil
}

func TestNewClient(t *testing.T){
	convey.Convey("given a new client instance is created", t, func() {
		actual := NewClient("baseurl", &broker.Broker{}, &http.Client{})
		convey.Convey("then a new client should be created", func() {
			convey.So(actual,convey.ShouldNotBeNil)
			convey.So(actual.baseurl, convey.ShouldEqual, "baseurl")
			convey.So(actual.broker, convey.ShouldResemble, &broker.Broker{})
			convey.So(actual.client, convey.ShouldResemble, &http.Client{})
		})
	})
}

//func TestPublishToBroker(t *testing.T){
//	convey.Convey("given a mock broker and http client is called", t, func() {
//		broker := &mockBroker{}
//		broker.On("Publish", mock.Anything).Return()
//		httpClient := &mockHttpClient{}
//		httpClient.On("GET", mock.Anything).Return(&http.Response{StatusCode:200,
//			Body: &mockBody{strings.NewReader("test message\n")},
//		}, nil)
//		client := NewClient("baseurl", broker, httpClient)
//		convey.Convey("when a new message is published", func(){
//		})
//	})
//}

//error handling if http client returns an error
