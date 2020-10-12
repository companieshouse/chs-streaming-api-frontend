package broker

//A broker to which producers will send messages published to all subscribed consumers.
type Broker struct {
	userSubscribed   chan *Event
	userUnsubscribed chan *Event
	users            map[chan string]bool
	data                 chan string
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
					msg:       "Attempted to unsubscribe a consumer that was not subscribed",
				}
				continue
			}
			delete(b.users, unsubscribed.stream)
			close(unsubscribed.stream)
			unsubscribed.result <- &Result{}
		case data := <-b.data:
			for consumer := range b.users {
				consumer <- data
			}
		}
	}
}

