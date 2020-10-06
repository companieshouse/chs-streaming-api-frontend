package main

import (
	"fmt"
	"github.com/companieshouse/chs-streaming-api-frontend/offset"
	"time"

	"github.com/companieshouse/chs-streaming-api-frontend/config"
	"github.com/companieshouse/chs-streaming-api-frontend/handlers"
	"github.com/companieshouse/chs-streaming-api-frontend/consumer"
	"github.com/companieshouse/chs-streaming-api-frontend/handlers/filing"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/chs.go/service"
	"github.com/companieshouse/chs.go/service/authorise"
	"github.com/companieshouse/chs.go/service/handlers/requestID"
	"os"

	"github.com/justinas/alice"
)


const (
	cacheBroker             = "cache-broker"
	filingHistoryStream     = "stream-filing-history"
	companyProfileStream    = "stream-company-profile"
	companyInsolvencyStream = "stream-company-insolvency"
	companyChargesStream    = "stream-company-charges"
	companyOfficersStream   = "stream-company-officers"
	companyPSCStream        = "stream-company-psc"
)

func main() {

	//check if config exists
	cfg, err := config.Get()
	if err != nil {
		log.Error(fmt.Errorf("error configuring service: %s. Exiting", err), nil)
		return
	}

	log.Namespace = cfg.Namespace()
	log.Info("initialising chs streaming api frontend service ...", log.Data{"config": cfg})

	service.DefaultMiddleware = []alice.Constructor{requestID.Handler(20), log.Handler}

	svc := service.New(cfg.ServiceConfig())

	//Authentication of incoming request
	auth := authorise.Authorisation{EricAddr: cfg.EricURL}
	svc.Chain(auth.Do)

	//service healthcheck
	svc.Router().Path("/healthcheck").Methods("GET").HandlerFunc(handlers.HealthCheck)


	//TODO : Call to cache broker
	log.Info("Getting cache-broker", log.Data{"cache-broker": cacheBroker})
	//s, err := cacheBroker.Get(cfg.CacheBrokerURL, cacheBroker)
	if err != nil {
		log.Error(fmt.Errorf("error receiving %s cacheBroker: %s", cacheBroker, err))
		os.Exit(1)
	}


	//Streaming Request
	streamHandler := handlers.Streaming{
		BrokerAddr:        cfg.StreamingBrokerAddr,
		RequestTimeout:    time.Duration(cfg.RequestTimeout),
		HeartbeatInterval: time.Duration(cfg.HeartbeatInterval),
		//CacheBroker:       s,
		ConsumerFactory:   consumer.NewConsumerFactory(),
		Offset:            offset.NewOffset(),
	}

	filingStream := &filing.Streaming{}
	companyStream := &filing.Streaming{}
	insolvencyStream := &filing.Streaming{}
	chargesStream := &filing.Streaming{}
	officersStream := &filing.Streaming{}
	pscStream := &filing.Streaming{}



	streamHandler.AddStream(svc.Router(), "/filings", filingHistoryStream, filingStream)
	streamHandler.AddStream(svc.Router(), "/companies", companyProfileStream, companyStream)
	streamHandler.AddStream(svc.Router(), "/insolvency-cases", companyInsolvencyStream, insolvencyStream)
	streamHandler.AddStream(svc.Router(), "/charges", companyChargesStream, chargesStream)
	streamHandler.AddStream(svc.Router(), "/officers", companyOfficersStream, officersStream)
	streamHandler.AddStream(svc.Router(), "/persons-with-significant-control", companyPSCStream, pscStream)


	svc.Start()
}

