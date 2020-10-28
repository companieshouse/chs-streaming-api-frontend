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
	GetTimer(duration time.Duration) *time.Timer
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

func (t *TimerFactory) GetTimer(duration time.Duration) *time.Timer {
	return time.NewTimer(duration * t.Unit)
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
