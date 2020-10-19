package broker

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateNewBrokerInstance(t *testing.T) {
	Convey("When a new cache broker instance is created", t, func() {
		actual := NewBroker()
		Convey("Then a new broker instance should be returned", func() {
			So(actual, ShouldNotBeNil)
		})
	})
}

func TestSubscribeUser(t *testing.T) {
	Convey("Given a running cache broker instance", t, func() {
		broker := NewBroker()
		go broker.Run()
		Convey("When a user subscribes to the cache broker", func() {
			user, err := broker.Subscribe()
			Convey("Then a new subscription should be created", func() {
				So(user, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(broker.users, ShouldContainKey, user)
			})
		})
	})
}

func TestUnsubscribeUser(t *testing.T) {
	Convey("Given a running broker instance with a subscribed consumer", t, func() {
		broker := NewBroker()
		go broker.Run()
		consumer, _ := broker.Subscribe()
		Convey("When the consumer unsubscribes", func() {
			err := broker.Unsubscribe(consumer)
			Convey("Then the user should be removed from the list of subscribers", func() {
				So(broker.users, ShouldNotContainKey, consumer)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestUnsubscribeUserReturnsErrorIfNotSubscribed(t *testing.T) {
	Convey("Given a running cache broker instance with no subscribed users", t, func() {
		broker := NewBroker()
		go broker.Run()
		Convey("When an unsubscribed user attempts to unsubscribe", func() {
			err := broker.Unsubscribe(make(chan string))
			Convey("Then an error should be returned", func() {
				So(err.Error(), ShouldEqual, "Attempted to unsubscribe a user that was not subscribed")
			})
		})
	})
}

func TestPublishMessage(t *testing.T) {
	Convey("Given a running broker instance with a subscribed user", t, func() {
		broker := NewBroker()
		go broker.Run()
		user, _ := broker.Subscribe()
		Convey("When a message is published", func() {
			broker.Publish("Hello world!")
			Convey("Then the message should be published to all subscribers", func() {
				So(<-user, ShouldEqual, "Hello world!")
			})
		})
	})
}
