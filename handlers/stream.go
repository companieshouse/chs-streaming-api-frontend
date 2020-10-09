package handlers

import (
	"net/http"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/gorilla/pat"
)

//Streaming contains necessary config for streaming
type Streaming struct {
	RequestTimeout    time.Duration
	HeartbeatInterval time.Duration
	//CacheBroker:       s,
}

// AddStream sets up the routing for the particular stream type
func (st Streaming) AddStream(router *pat.Router, route string, streamName string) {
	router.Path(route).Methods("GET").HandlerFunc(st.process(streamName))
}

func (st Streaming) process(streamName string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		log.InfoC(req.Header.Get("ERIC_Identity"), "consuming from cache-broker", log.Data{"Stream Name": streamName})

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

	//TODO The frontend will now be decoupled from Kafka, this handler should do two things:
	//		a. Fetch a range of offsets cached by chs-streaming-api-cache if a timepoint has been specified.
	//		b. Stream offsets published to it by chs-streaming-api-cache to the connected user.

	contextID := req.Header.Get("ERIC_Identity")
	log.DebugC(contextID, "Getting offset from the cache broker", log.Data{"timepoint": 10})

	heathcheckTimer := time.NewTimer(st.HeartbeatInterval * time.Second)
	requestTimer := time.NewTimer(st.RequestTimeout * time.Second)

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
		}
		//TODO when receiving message from cache-broker then process the message and writing back to response
		// 	   writer for connected users
	}
}

func responseWriterWrite(w http.ResponseWriter, b []byte) (int, error) {
	return w.Write(b)
}
