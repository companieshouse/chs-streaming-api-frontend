package broker

import "errors"

//A broker to which cache broker will send messages published to all subscribed users.
type Broker struct {
	userSubscribed   chan *Event
	userUnsubscribed chan *Event
	users            map[chan string]bool
	data             chan string
}

//An event that has been emitted to the given broker instance.
type Event struct {
	stream chan string
	result chan *Result
}

//The result of the event after it has been handled by the event handler.
type Result struct {
	hasErrors bool
	msg       string
}

//Create a new broker instance.
func NewBroker() *Broker {
	return &Broker{
		userSubscribed:   make(chan *Event),
		userUnsubscribed: make(chan *Event),
		users:            make(map[chan string]bool),
		data:             make(chan string),
	}
}

//Subscribe a user to this broker.
func (b *Broker) Subscribe() (chan string, error) {
	stream := make(chan string)
	subscription := &Event{
		stream: stream,
		result: make(chan *Result),
	}
	b.userSubscribed <- subscription
	<-subscription.result
	close(subscription.result)
	return stream, nil
}

//Run this broker instance.
func (b *Broker) Run() {
	for {
		select {
		case subscriber := <-b.userSubscribed:
			b.users[subscriber.stream] = true
			subscriber.result <- &Result{}
		case unsubscribed := <-b.userUnsubscribed:
			if _, ok := b.users[unsubscribed.stream]; !ok {
				unsubscribed.result <- &Result{
					hasErrors: true,
					msg:       "Attempted to unsubscribe a user that was not subscribed",
				}
				continue
			}
			delete(b.users, unsubscribed.stream)
			close(unsubscribed.stream)
			unsubscribed.result <- &Result{}
		case data := <-b.data:
			for user := range b.users {
				user <- data
			}
		}
	}
}

//Unsubscribe a user from this broker.
//If the user isn't subscribed to this broker then an error will be returned.
func (b *Broker) Unsubscribe(consumer chan string) error {
	subscription := &Event{
		stream: consumer,
		result: make(chan *Result),
	}
	defer close(subscription.result)
	b.userUnsubscribed <- subscription
	result := <-subscription.result
	if result.hasErrors {
		return errors.New(result.msg)
	}
	return nil
}

//Publish a message to all subscribed users.
func (b *Broker) Publish(msg string) {
	b.data <- msg
}