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

type Subscribable interface {
	Subscribe() (chan string, error)
	Unsubscribe(chan string) error
}

type TimerFactory struct {
	unit time.Duration
}

type ClientFactory struct {

}

func (c *ClientFactory) GetClient(baseurl string, path string, publisher client.Publishable, logger logger.Logger) *client.Client {
	return client.NewClient(baseurl, path, publisher, http.DefaultClient, logger)
}

func (t *TimerFactory) GetTimer(duration time.Duration) *time.Timer {
	return time.NewTimer(duration * t.unit)
}

type TimestampGeneratable interface {
	GetTimer(duration time.Duration) *time.Timer
}

//Streaming contains necessary config for streaming
type Streaming struct {
	RequestTimeout    time.Duration
	HeartbeatInterval time.Duration
	wg                *sync.WaitGroup
	Logger            logger.Logger
	CacheBrokerURL    string
	timerFactory      TimestampGeneratable
	clientFactory     ClientFactory
}

type Publisher struct {
	data chan string
}

func (p *Publisher) Publish(msg string) {
	p.data <- msg
}

/*
  Have mini functions for handling timer events so that we can stub these in tests to verify that they are called
*/
func handleHeartbeatTimeout(contextID string) {
	log.TraceC(contextID, "application heartbeat")
}

var callHeartbeatTimeout = handleHeartbeatTimeout

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
func (st Streaming) HandleRequest(streamName string, broker Subscribable, route string) func(writer http.ResponseWriter, req *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {

		if req.URL.Query().Get("timepoint") != "" {
			st.Logger.InfoR(req, "consuming from cache-broker, timepoint specified", log.Data{"Stream Name": streamName})
			st.ProcessOffsetHTTP(writer, req, route)
		} else {
			st.Logger.InfoR(req, "consuming from cache-broker", log.Data{"Stream Name": streamName})
			st.ProcessHTTP(writer, req, broker)
		}
	}
}

func (st Streaming) ProcessOffsetHTTP(writer http.ResponseWriter, request *http.Request, route string) {
	contextID := request.Header.Get("ERIC_Identity")
	heathcheckTimer := time.NewTimer(st.HeartbeatInterval * time.Second)
	requestTimer := time.NewTimer(st.RequestTimeout * time.Second)

	data := make(chan string)
	publisher := &Publisher{data}
	client2 := client.NewClient(st.CacheBrokerURL, route, publisher, http.DefaultClient, st.Logger)
	client2.Offset = request.URL.Query().Get("timepoint")

	go client2.Connect()

	for {
		select {
		case <-requestTimer.C:
			log.DebugC(contextID, "connection expired, disconnecting client ")
			return
		case <-request.Context().Done():
			close(data)
			log.DebugC(contextID, "connection closed by client")
			if st.wg != nil {
				st.wg.Done()
			}
			return
		case <-heathcheckTimer.C:
			heathcheckTimer.Reset(st.HeartbeatInterval * time.Second)

			writer.(http.Flusher).Flush()
			callHeartbeatTimeout(contextID)

			heathcheckTimer.Reset(st.HeartbeatInterval * time.Second)
			writer.(http.Flusher).Flush()
		case msg := <-data:
			st.Logger.InfoR(request, "User connected")
			_, _ = writer.Write([]byte(msg))
			writer.(http.Flusher).Flush()
			if st.wg != nil {
				st.wg.Done()
			}
		}
	}
}

func (st Streaming) ProcessHTTP(writer http.ResponseWriter, request *http.Request, broker Subscribable) {

	heartbeatTimer := st.timerFactory.GetTimer(st.HeartbeatInterval)
	requestTimer := st.timerFactory.GetTimer(st.RequestTimeout)

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

func responseWriterWrite(writer http.ResponseWriter, b []byte) (int, error) {
	return writer.Write(b)
}
