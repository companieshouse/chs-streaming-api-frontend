package broker


import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateNewBrokerInstance(t *testing.T) {
	Convey("When a new broker instance is created", t, func() {
		actual := NewBroker()
		Convey("Then a new broker instance should be returned", func() {
			So(actual, ShouldNotBeNil)
		})
	})
}

func TestSubscribeUser(t *testing.T) {
	Convey("Given a running broker instance", t, func() {
		broker := NewBroker()
		go broker.Run()
		Convey("When a user subscribes to the broker", func() {
			user, err := broker.Subscribe()
			Convey("Then a new subscription should be created", func() {
				So(user, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(broker.users, ShouldContainKey, user)
			})
		})
	})
}



