package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Shopify/sarama"
	"github.com/companieshouse/chs-go-avro-schemas/data"
	"github.com/companieshouse/chs-streaming-api-frontend/consumer"
	"github.com/companieshouse/chs-streaming-api-frontend/offset"

	"github.com/companieshouse/chs.go/log"
	"github.com/gorilla/pat"
)

// Stream should be implemented by the different stream types
type Stream interface {
	// Transform modifies data if required within this microservice
	Transform(req *http.Request, msg *[]byte) error
	// Filter modifies the data to a given filtered view set in the request
	Filter(req *http.Request, msg *[]byte) error
}

//Streaming contains necessary config for streaming
type Streaming struct {
	BrokerAddr        []string
	RequestTimeout    time.Duration
	HeartbeatInterval time.Duration
	//CacheBroker:       s,
	ConsumerFactory   consumer.FactoryInterface
	Offset            offset.Interface
}

// AddStream sets up the routing for the particular stream type
func (st Streaming) AddStream(router *pat.Router, route string, topic string, stream Stream) {
	router.Path(route).Methods("GET").HandlerFunc(st.process(topic, stream))
}

func (st Streaming) process(topic string, stream Stream) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		log.InfoC(req.Header.Get("ERIC_Identity"), "consuming from topic", log.Data{"topic": topic})

		st.ProcessHTTP(topic, w, req, stream)
	}
}

/*
  Have mini functions for handling timer events so that we can stub these in tests to verify that they are called
*/
func handleTimeOut(contextID string) {
	log.DebugC(contextID, "connection expired, disconnecting client ")
}

func handleClientDisconnect(contextID string) {
	log.DebugC(contextID, "connection closed by client")
}

func logHeartbeatTimeout(contextID string) {
	log.TraceC(contextID, "application heartbeat")
}

var callLogHandleTimeOut = handleTimeOut
var callLogHandleClientDisconnect = handleClientDisconnect
var callProcessMessage = ProcessMessage
var callProcessMessageFailed = processMessageFailed
var callLogHeartbeatTimeout = logHeartbeatTimeout
var callCheckIfDataNotPresent = checkIfDataNotPresent
var callResponseWriterWrite = responseWriterWrite

// ProcessHTTP Test at this level for unit testing.
func (st Streaming) ProcessHTTP(topic string, w http.ResponseWriter, req *http.Request, stream Stream) {

	o, err := st.Offset.Parse(req.URL.Query().Get("timepoint"))

	contextID := req.Header.Get("ERIC_Identity")
	log.DebugC(contextID, "Getting offset from the url", log.Data{"timepoint": o, "topic": topic})

	if err != nil {
		log.DebugC(contextID, "invalid offset requested", log.Data{"error": err, "offset": o})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//if no timepoint is added then offset is set to newest
	if o != 0 {
		if err = st.Offset.IsValid(st.BrokerAddr, topic, o); err != nil {
			if err == offset.ErrOutOfRange {
				log.DebugC(contextID, "requested offset out of range", log.Data{"error": err, "offset": o})
				w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				return
			}
			log.ErrorC(contextID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		o = sarama.OffsetNewest
		log.DebugC(contextID, "No Timepoint set, set offset to newest offset ", log.Data{"offset": o})
	}

	con := st.ConsumerFactory.Initialise(topic, st.BrokerAddr)
	if err := con.ConsumePartition(0, o); err != nil {
		log.ErrorC(contextID, fmt.Errorf("error consuming from topic %s at offset %d: '%s'", topic, o, err))

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hbTimer := time.NewTimer(st.HeartbeatInterval * time.Second)
	huTimer := time.NewTimer(st.RequestTimeout * time.Second)
	msgChan := con.Messages()

	for {
		select {
		case <-huTimer.C:
			callLogHandleTimeOut(contextID)
			return
		case <-hbTimer.C:
			hbTimer.Reset(st.HeartbeatInterval * time.Second)

			if _, err = w.Write([]byte("\n")); err != nil {
				log.ErrorR(req, err)
				continue
			}

			w.(http.Flusher).Flush()
			callLogHeartbeatTimeout(contextID)
		case m := <-msgChan:

			err = callProcessMessage(contextID, m, w, stream)
			if err != nil {
				callProcessMessageFailed(contextID, err)
				continue
			}

			hbTimer.Reset(st.HeartbeatInterval * time.Second)
			w.(http.Flusher).Flush()
		}
	}
}

func checkIfDataNotPresent(sch data.ResourceChangedData) error {
	if len(sch.Data) == 0 {
		return errors.New("No data is available for the stream")
	}
	return nil
}

func processMessageFailed(contextID string, err error) {
	log.DebugC(contextID, "error returned from consuming a message : '%s'", log.Data{"err": err})
}

func responseWriterWrite(w http.ResponseWriter, b []byte) (int, error) {
	return w.Write(b)
}

// ProcessMessage - This method is processingvalidating the message and writing the message back to response writer
func ProcessMessage(contextID string, m *sarama.ConsumerMessage, w http.ResponseWriter, stream Stream) error {

	var err error
	var resourceChangedData data.ResourceChangedData

	if err = callCheckIfDataNotPresent(resourceChangedData); err != nil {
		log.ErrorC(contextID, fmt.Errorf("Data is not available for processing: '%s'", err))
		return err
	}

	log.TraceC(contextID, resourceChangedData.ResourceID)


	if _, err = callResponseWriterWrite(w, []byte(fmt.Sprintf("%s\n"))); err != nil {
		log.ErrorC(contextID, fmt.Errorf("fail to write header: '%s'", err))
		return err
	}

	return err

}
