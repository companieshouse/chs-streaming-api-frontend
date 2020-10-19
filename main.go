package main

import (
	"fmt"
	"github.com/companieshouse/chs-streaming-api-frontend/broker"
	"github.com/companieshouse/chs-streaming-api-frontend/config"
	"github.com/companieshouse/chs-streaming-api-frontend/handlers"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/chs.go/service"
	"github.com/companieshouse/chs.go/service/authorise"
	"github.com/companieshouse/chs.go/service/handlers/requestID"

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

	//connect to cache-broker
	log.Info("Getting cache-broker", log.Data{"cache-broker": cacheBroker})
	publisher := broker.NewBroker() //incoming messages

	//Streaming Request Handler
	streamHandler := handlers.Streaming{
		RequestTimeout:    time.Duration(cfg.RequestTimeout),
		HeartbeatInterval: time.Duration(cfg.HeartbeatInterval),
		Broker:            publisher,
	}

	isOfficersEnabled := cfg.OfficersEndpointFlag
	isPSCsEnabled := cfg.PSCsEndpointFlag

	streamHandler.AddStream(svc.Router(), "/filings", filingHistoryStream)
	streamHandler.AddStream(svc.Router(), "/companies", companyProfileStream)
	streamHandler.AddStream(svc.Router(), "/insolvency-cases", companyInsolvencyStream)
	streamHandler.AddStream(svc.Router(), "/charges", companyChargesStream)

	if isOfficersEnabled {
		streamHandler.AddStream(svc.Router(), "/officers", companyOfficersStream)
	}
	if isPSCsEnabled {
		streamHandler.AddStream(svc.Router(), "/persons-with-significant-control", companyPSCStream)
	}

	svc.Start()
}
