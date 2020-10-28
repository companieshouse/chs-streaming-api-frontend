package factory

import (
	"github.com/companieshouse/chs-streaming-api-frontend/broker"
	"github.com/companieshouse/chs-streaming-api-frontend/logger"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestGetClientReturnsNewClientInstance(t *testing.T) {
	Convey("Given a new client factory instance has been created", t, func() {
		factory := &ClientFactory{}
		publisher := &Publisher{}
		loggerInstance := &logger.LoggerImpl{}
		Convey("When a new client instance is obtained", func() {
			actual := factory.GetClient("baseurl", "/path", publisher, loggerInstance)
			Convey("Then a client instance constructed from the given params should be returned", func() {
				So(actual, ShouldNotBeNil)
			})
		})
	})
}

func TestGetTimerReturnsNewTimerInstance(t *testing.T) {
	Convey("Given a new timer factory instance has been created with a time unit of seconds", t, func() {
		factory := &TimerFactory{Unit: time.Second}
		Convey("When a new timer client instance is obtained", func() {
			actual := factory.GetTimer(0)
			Convey("Then a timer with a timeout of duration * time unit should be created", func() {
				So(<-actual.C, ShouldNotBeNil)
			})
		})
	})
}

func TestGetPublisherReturnsNewPublisherInstance(t *testing.T) {
	Convey("Given a new publisher factory instance has been created", t, func() {
		factory := &PublisherFactory{}
		Convey("When a new publisher instance is obtained", func() {
			actual := factory.GetPublisher()
			Convey("Then a new publisher instance should be returned", func() {
				So(actual, ShouldHaveSameTypeAs, &Publisher{})
			})
		})
	})
}

func TestSubscribeToPublisherReturnsChannel(t *testing.T) {
	Convey("Given a new publisher instance has been created", t, func() {
		dataChannel := make(chan string)
		publisher := &Publisher{data: dataChannel}
		Convey("When the publisher is subscribed to", func() {
			actual, err := publisher.Subscribe()
			Convey("Then the channel associated to the publisher should be returned", func() {
				So(actual, ShouldEqual, dataChannel)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestUnsubscribeFromPublisher(t *testing.T) {
	Convey("Given a new publisher instance has been created and a user has subscribed to it", t, func() {
		dataChannel := make(chan string)
		publisher := &Publisher{data: dataChannel}
		subscription, _ := publisher.Subscribe()
		Convey("When the user unsubscribes from the publisher", func() {
			err := publisher.Unsubscribe(subscription)
			Convey("Then the channel will be closed and no errors should be returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestUnsubscribeFromPublisherReturnsErrorIfChannelIsNotAssociated(t *testing.T) {
	Convey("Given a new publisher instance has been created and a user has subscribed to it", t, func() {
		dataChannel := make(chan string)
		otherChannel := make(chan string)
		close(otherChannel)
		publisher := &Publisher{data: dataChannel}
		Convey("When the user unsubscribes from the publisher using a different subscription", func() {
			err := publisher.Unsubscribe(make(chan string))
			Convey("Then the channel will remain open and an error should be returned", func() {
				So(err.Error(), ShouldEqual, "Attempted to close an unmanaged channel")
			})
		})
	})
}

func TestPublishToPublisher(t *testing.T) {
	Convey("Given a new publisher instance has been created and a user has subscribed to it", t, func() {
		dataChannel := make(chan string)
		publisher := &Publisher{data: dataChannel}
		actual, _ := publisher.Subscribe()
		Convey("When a producer in a different goroutine publishes a message", func() {
			go func() { publisher.Publish("Hello world") }()
			Convey("Then the message should be consumed by the subscriber", func() {
				So(<-actual, ShouldEqual, "Hello world")
			})
		})
	})
}

func TestGetBrokerReturnsNewBrokerInstance(t *testing.T) {
	Convey("Given a new publisher factory instance has been created", t, func() {
		factory := &PublisherFactory{}
		Convey("When a new broker instance is obtained", func() {
			actual := factory.GetBroker()
			Convey("Then a new broker instance should be returned", func() {
				So(actual, ShouldHaveSameTypeAs, &broker.CacheBroker{})
			})
		})
	})
}
