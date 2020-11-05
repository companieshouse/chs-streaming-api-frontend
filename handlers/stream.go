package handlers

import (
	"github.com/companieshouse/chs-streaming-api-frontend/broker"
	"github.com/companieshouse/chs-streaming-api-frontend/client"
	"github.com/companieshouse/chs-streaming-api-frontend/factory"
	"github.com/companieshouse/chs-streaming-api-frontend/logger"
	"github.com/companieshouse/chs.go/log"

	"github.com/gorilla/pat"
	"net/http"
	"sync"
	"time"
)

//Streaming contains necessary config for streaming
type Streaming struct {
	RequestTimeout    time.Duration
	HeartbeatInterval time.Duration
	wg                *sync.WaitGroup
	Logger            logger.Logger
	CacheBrokerURL    string
	TimerFactory      factory.TimestampGeneratable
	ClientFactory     factory.ClientGettable
	PublisherFactory  factory.PublisherGettable
	ApiKey            string
}

// AddStream sets up the routing for the particular stream type
func (st Streaming) AddStream(router *pat.Router, backendPath string, route string, streamName string) {
	broker := broker.NewBroker() //incoming messages
	//connect to cache-broker
	client2 := st.ClientFactory.GetClient(st.CacheBrokerURL, backendPath, st.ApiKey, broker, st.Logger)
	client2.Connect()
	go broker.Run()

	router.Path(route).Methods("GET").HandlerFunc(st.HandleRequest(streamName, broker, backendPath))
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
	var serviceClient factory.Connectable

	requestTimer := st.TimerFactory.GetTimer(st.RequestTimeout * time.Second)
	heartbeatTimer := st.TimerFactory.GetTimer(st.HeartbeatInterval * time.Second)

	st.Logger.Info("Handling offset")
	if route != "" {
		serviceClient = st.ClientFactory.GetClient(st.CacheBrokerURL, route, st.ApiKey, broker, st.Logger)
		serviceClient.SetOffset(request.URL.Query().Get("timepoint"))
		serviceClient.Connect()
	}
	st.Logger.Info("Subscribing to broker")
	subscription, _ := broker.Subscribe()

	requestTimer.Start()
	heartbeatTimer.Start()

	for {
		select {
		case <-requestTimer.Elapsed():
			requestTimer.Stop()
			heartbeatTimer.Stop()
			if serviceClient != nil {
				serviceClient.Close()
			}
			_ = broker.Unsubscribe(subscription)
			log.InfoR(request, "connection expired, disconnecting client ")
			if st.wg != nil {
				st.wg.Done()
			}
			return
		case <-request.Context().Done():
			requestTimer.Stop()
			heartbeatTimer.Stop()
			if serviceClient != nil {
				serviceClient.Close()
			}
			_ = broker.Unsubscribe(subscription)
			st.Logger.InfoR(request, "User disconnected")
			if st.wg != nil {
				st.wg.Done()
			}
			return
		case <-heartbeatTimer.Elapsed():
			heartbeatTimer.Reset()
			_, _ = writer.Write([]byte("\n"))
			writer.(http.Flusher).Flush()
			st.Logger.InfoR(request, "Application heartbeat")
			if st.wg != nil {
				st.wg.Done()
			}
		case msg := <-subscription:
			_, _ = writer.Write([]byte(msg))
			writer.(http.Flusher).Flush()
			if st.wg != nil {
				st.wg.Done()
			}
		}
	}
}
