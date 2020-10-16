package handlers

import (
	"github.com/companieshouse/chs-streaming-api-cache/logger"
	"github.com/companieshouse/chs.go/log"

	"github.com/gorilla/pat"
	"net/http"
	"sync"
	"time"
)

type CacheBroker interface {
	Subscribe() (chan string, error)
	Unsubscribe(chan string) error
}

//Streaming contains necessary config for streaming
type Streaming struct {
	RequestTimeout    time.Duration
	HeartbeatInterval time.Duration
	Broker            CacheBroker
	wg                *sync.WaitGroup
	logger            logger.Logger
}

// AddStream sets up the routing for the particular stream type
func (st Streaming) AddStream(router *pat.Router, route string, streamName string) {
	router.Path(route).Methods("GET").HandlerFunc(st.process(streamName))
}

func (st Streaming) process(streamName string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		st.logger.InfoR(req, "consuming from cache-broker", log.Data{"Stream Name": streamName})
		st.ProcessHTTP(w, req)
	}
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

func (st Streaming) ProcessHTTP(w http.ResponseWriter, req *http.Request) {

	contextID := req.Header.Get("ERIC_Identity")
	heathcheckTimer := time.NewTimer(st.HeartbeatInterval * time.Second)
	requestTimer := time.NewTimer(st.RequestTimeout * time.Second)

	subscription, _ := st.Broker.Subscribe()

	for {
		select {
		case <-requestTimer.C:
			callRequestTimeOut(contextID)
			return
		case <-req.Context().Done():
			callHandleClientDisconnect(contextID)
			return
		case <-heathcheckTimer.C:
			heathcheckTimer.Reset(st.HeartbeatInterval * time.Second)

			w.(http.Flusher).Flush()
			callHeartbeatTimeout(contextID)

			heathcheckTimer.Reset(st.HeartbeatInterval * time.Second)
			w.(http.Flusher).Flush()
		case msg := <-subscription:
			st.logger.InfoR(req, "User connected")
			_, _ = w.Write([]byte(msg))
			w.(http.Flusher).Flush()
			if st.wg != nil {
				st.wg.Done()
			}
		case <-req.Context().Done():
			_ = st.Broker.Unsubscribe(subscription)
			st.logger.InfoR(req, "User disconnected")
			if st.wg != nil {
				st.wg.Done()
			}
			return
		}

	}
}

func responseWriterWrite(w http.ResponseWriter, b []byte) (int, error) {
	return w.Write(b)
}
