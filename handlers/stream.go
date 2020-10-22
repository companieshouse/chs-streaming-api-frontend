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

//Streaming contains necessary config for streaming
type Streaming struct {
	RequestTimeout    time.Duration
	HeartbeatInterval time.Duration
	wg                *sync.WaitGroup
	Logger            logger.Logger
}

/*
  Have mini functions for handling timer events so that we can stub these in tests to verify that they are called
*/
func handleRequestTimeOut(contextID string) {
	log.DebugC(contextID, "connection expired, disconnecting client ")
}

func handleClientDisconnect(contextID string) {
	log.DebugC(contextID, "connection closed by client")
}

func handleHeartbeatTimeout(contextID string) {
	log.TraceC(contextID, "application heartbeat")
}

var callResponseWriterWrite = responseWriterWrite

var callRequestTimeOut = handleRequestTimeOut
var callHeartbeatTimeout = handleHeartbeatTimeout
var callHandleClientDisconnect = handleClientDisconnect

// AddStream sets up the routing for the particular stream type
func (st Streaming) AddStream(router *pat.Router, route string, streamName string, cacheBrokerUrl string) {

	broker := broker.NewBroker() //incoming messages
	//connect to cache-broker
	client2 := client.NewClient(cacheBrokerUrl, broker, http.DefaultClient, st.Logger)
	go client2.Connect()
	go broker.Run()

	router.Path(route).Methods("GET").HandlerFunc(st.HandleRequest(streamName, broker))
}

// Handle Request
func (st Streaming) HandleRequest(streamName string, broker Subscribable) func(writer http.ResponseWriter, req *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {

		st.Logger.InfoR(req, "consuming from cache-broker", log.Data{"Stream Name": streamName})
		st.ProcessHTTP(writer, req, broker)
	}
}

func (st Streaming) ProcessHTTP(writer http.ResponseWriter, request *http.Request, broker Subscribable) {

	contextID := request.Header.Get("ERIC_Identity")
	heathcheckTimer := time.NewTimer(st.HeartbeatInterval * time.Second)
	requestTimer := time.NewTimer(st.RequestTimeout * time.Second)

	subscription, _ := broker.Subscribe()

	for {
		select {
		case <-requestTimer.C:
			callRequestTimeOut(contextID)
			return
		case <-request.Context().Done():
			callHandleClientDisconnect(contextID)
			return
		case <-heathcheckTimer.C:
			heathcheckTimer.Reset(st.HeartbeatInterval * time.Second)

			writer.(http.Flusher).Flush()
			callHeartbeatTimeout(contextID)

			heathcheckTimer.Reset(st.HeartbeatInterval * time.Second)
			writer.(http.Flusher).Flush()
		case msg := <-subscription:
			st.Logger.InfoR(request, "User connected")
			_, _ = writer.Write([]byte(msg))
			writer.(http.Flusher).Flush()
			if st.wg != nil {
				st.wg.Done()
			}
		case <-request.Context().Done():
			_ = broker.Unsubscribe(subscription)
			st.Logger.InfoR(request, "User disconnected")
			if st.wg != nil {
				st.wg.Done()
			}
			return
		}

	}
}

func responseWriterWrite(writer http.ResponseWriter, b []byte) (int, error) {
	return writer.Write(b)
}
