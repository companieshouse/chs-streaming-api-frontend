package factory

import (
	"errors"
	"github.com/companieshouse/chs-streaming-api-frontend/broker"
	"github.com/companieshouse/chs-streaming-api-frontend/client"
	"github.com/companieshouse/chs-streaming-api-frontend/logger"
	"net/http"
	"time"
)

type TimestampGeneratable interface {
	GetTimer(duration time.Duration) Elapseable
}

type Connectable interface {
	Connect() *client.ResponseStatus
	SetOffset(offset string)
	Close()
}

type RunnablePublisher interface {
	client.Publishable
	Run()
}

type ClientGettable interface {
	GetClient(baseurl string, path string, publisher client.Publishable, logger logger.Logger) Connectable
}

type PublisherGettable interface {
	GetPublisher() client.Publishable
	GetBroker() RunnablePublisher
}

type Elapseable interface {
	Start()
	Elapsed() <-chan bool
	Reset()
	Stop()
}

type Interval struct {
	started       bool
	notifications chan bool
	pulse         chan bool
	interval      time.Duration
}

func (i *Interval) Start() {
	go func() {
		t := time.NewTimer(0)
		<-t.C
		for <-i.pulse {
			t.Reset(i.interval)
			<-t.C
			select {
			case i.notifications <- true:
			default:
			}
		}
		t.Stop()
	}()
	i.started = true
	i.pulse <- true
}

func (i *Interval) Elapsed() <-chan bool {
	return i.notifications
}

func (i *Interval) Reset() {
	select {
	case i.pulse <- true:
	default:
	}
}

func (i *Interval) Stop() {
	select {
	case i.pulse <- false:
	default:
	}
	i.started = false
}

type TimerFactory struct {
	Unit time.Duration
}

type ClientFactory struct {
}

type PublisherFactory struct {
}

type Publisher struct {
	data chan string
}

func (c *ClientFactory) GetClient(baseurl string, path string, publisher client.Publishable, logger logger.Logger) Connectable {
	return client.NewClient(baseurl, path, publisher, http.DefaultClient, logger)
}

func (t *TimerFactory) GetTimer(duration time.Duration) Elapseable {
	return &Interval{
		interval:      duration,
		notifications: make(chan bool),
		pulse:         make(chan bool),
	}
}

func (p *PublisherFactory) GetPublisher() client.Publishable {
	return &Publisher{data: make(chan string)}
}

func (p *PublisherFactory) GetBroker() RunnablePublisher {
	return broker.NewBroker()
}

func (p *Publisher) Publish(msg string) {
	p.data <- msg
}

func (p *Publisher) Subscribe() (chan string, error) {
	return p.data, nil
}

func (p *Publisher) Unsubscribe(subscription chan string) error {
	if p.data != subscription {
		return errors.New("Attempted to close an unmanaged channel")
	}
	close(subscription)
	return nil
}
