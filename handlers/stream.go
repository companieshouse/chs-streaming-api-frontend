package handlers

import (
	"github.com/companieshouse/chs-streaming-api-frontend/broker"
	"github.com/companieshouse/chs-streaming-api-frontend/client"
	"github.com/companieshouse/chs-streaming-api-frontend/logger"
	"github.com/companieshouse/chs.go/log"

	"github.com/gorilla/pat"
	"net/http"
	"sync"
	"time"
)

type TimerFactory struct {
	Unit time.Duration
}

type ClientFactory struct {
}

type PublisherFactory struct {
}

func (c *ClientFactory) GetClient(baseurl string, path string, publisher client.Publishable, logger logger.Logger) Connectable {
	return client.NewClient(baseurl, path, publisher, http.DefaultClient, logger)
}

func (t *TimerFactory) GetTimer(duration time.Duration) *time.Timer {
	return time.NewTimer(duration * t.Unit)
}

func (p *PublisherFactory) GetPublisher() client.Publishable {
	return &Publisher{}
}

type TimestampGeneratable interface {
	GetTimer(duration time.Duration) *time.Timer
}

type Connectable interface {
	Connect()
	SetOffset(offset string)
}

type ClientGettable interface {
	GetClient(baseurl string, path string, publisher client.Publishable, logger logger.Logger) Connectable
}

type PublisherGettable interface {
	GetPublisher() client.Publishable
}

//Streaming contains necessary config for streaming
type Streaming struct {
	RequestTimeout    time.Duration
	HeartbeatInterval time.Duration
	wg                *sync.WaitGroup
	Logger            logger.Logger
	CacheBrokerURL    string
	TimerFactory      TimestampGeneratable
	ClientFactory     ClientGettable
	PublisherFactory  PublisherGettable
}

type Publisher struct {
	data chan string
}

func (p *Publisher) Publish(msg string) {
	p.data <- msg
}

func (p *Publisher) Subscribe() (chan string, error) {
	p.data = make(chan string)
	return p.data, nil
}

func (p *Publisher) Unsubscribe(subscription chan string) error {
	close(subscription)
	return nil
}

// AddStream sets up the routing for the particular stream type
func (st Streaming) AddStream(router *pat.Router, route string, streamName string) {

	broker := broker.NewBroker() //incoming messages
	//connect to cache-broker
	client2 := client.NewClient(st.CacheBrokerURL, route, broker, http.DefaultClient, st.Logger)
	go client2.Connect()
	go broker.Run()

	router.Path(route).Methods("GET").HandlerFunc(st.HandleRequest(streamName, broker, route))
}

// Handle Request
func (st Streaming) HandleRequest(streamName string, broker client.Publishable, route string) func(writer http.ResponseWriter, req *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {

		if req.URL.Query().Get("timepoint") != "" {
			st.Logger.InfoR(req, "consuming from cache-broker, timepoint specified", log.Data{"Stream Name": streamName})
			st.ProcessHTTP(writer, req, route, st.PublisherFactory.GetPublisher())
		} else {
			st.Logger.InfoR(req, "consuming from cache-broker", log.Data{"Stream Name": streamName})
			st.ProcessHTTP(writer, req, "", broker)
		}
	}
}

func (st Streaming) ProcessHTTP(writer http.ResponseWriter, request *http.Request, route string, broker client.Publishable) {

	heartbeatTimer := st.TimerFactory.GetTimer(st.HeartbeatInterval)
	requestTimer := st.TimerFactory.GetTimer(st.RequestTimeout)

	if route != "" {
		client2 := st.ClientFactory.GetClient(st.CacheBrokerURL, route, broker, st.Logger)
		client2.SetOffset(request.URL.Query().Get("timepoint"))
		go client2.Connect()
	}

	subscription, _ := broker.Subscribe()

	for {
		select {
		case <-requestTimer.C:
			_ = broker.Unsubscribe(subscription)
			log.InfoR(request, "connection expired, disconnecting client ")
			if st.wg != nil {
				st.wg.Done()
			}
			return
		case <-request.Context().Done():
			_ = broker.Unsubscribe(subscription)
			st.Logger.InfoR(request, "User disconnected")
			if st.wg != nil {
				st.wg.Done()
			}
			return
		case <-heartbeatTimer.C:
			heartbeatTimer.Reset(st.HeartbeatInterval * time.Second)
			_, _ = writer.Write([]byte("\n"))
			writer.(http.Flusher).Flush()
			st.Logger.InfoR(request, "Application heartbeat")
			if st.wg != nil {
				st.wg.Done()
			}
		case msg := <-subscription:
			st.Logger.InfoR(request, "User connected")
			_, _ = writer.Write([]byte(msg))
			writer.(http.Flusher).Flush()
			if st.wg != nil {
				st.wg.Done()
			}
		}
	}
}
