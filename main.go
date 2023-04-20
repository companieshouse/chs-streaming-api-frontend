package main

import (
	"fmt"
	"time"

	"github.com/companieshouse/chs-streaming-api-frontend/config"
	"github.com/companieshouse/chs-streaming-api-frontend/factory"
	"github.com/companieshouse/chs-streaming-api-frontend/handlers"
	"github.com/companieshouse/chs-streaming-api-frontend/logger"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/chs.go/service"
	"github.com/companieshouse/chs.go/service/authorise"
	"github.com/companieshouse/chs.go/service/handlers/requestID"

	"github.com/justinas/alice"
)

const (
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

	log.Info("cfg.ApiKey", log.Data{"cfg.ApiKey": cfg.ApiKey})

	service.DefaultMiddleware = []alice.Constructor{requestID.Handler(20), log.Handler}

	svc := service.New(cfg.ServiceConfig())

	//Authentication of incoming request
	auth := authorise.Authorisation{EricAddr: cfg.EricURL}
	svc.Chain(auth.Do)

	//service healthcheck
	svc.Router().Path("/healthcheck").Methods("GET").HandlerFunc(handlers.HealthCheck)

	//Streaming Request Handler
	streamHandler := handlers.Streaming{
		RequestTimeout:    time.Duration(cfg.RequestTimeout),
		HeartbeatInterval: time.Duration(cfg.HeartbeatInterval),
		Logger:            logger.NewLogger(),
		CacheBrokerURL:    cfg.CacheBrokerURL,
		ClientFactory:     &factory.ClientFactory{},
		PublisherFactory:  &factory.PublisherFactory{},
		TimerFactory:      &factory.TimerFactory{},
		ApiKey:            cfg.ApiKey,
	}

	//Get stream enabled flag values from config
	isOfficersEnabled := cfg.OfficersEndpointFlag
	isPSCsEnabled := cfg.PSCsEndpointFlag

	streamHandler.AddStream(svc.Router(), "/streaming-api-cache/filings", "/filings", filingHistoryStream)
	streamHandler.AddStream(svc.Router(), "/streaming-api-cache/companies", "/companies", companyProfileStream)
	streamHandler.AddStream(svc.Router(), "/streaming-api-cache/insolvency-cases", "/insolvency-cases", companyInsolvencyStream)
	streamHandler.AddStream(svc.Router(), "/streaming-api-cache/charges", "/charges", companyChargesStream)

	if isOfficersEnabled {
		streamHandler.AddStream(svc.Router(), "/streaming-api-cache/officers", "/officers", companyOfficersStream)
	}
	if isPSCsEnabled {
		streamHandler.AddStream(svc.Router(), "/streaming-api-cache/persons-with-significant-control", "/persons-with-significant-control", companyPSCStream)
	}

	svc.Start()
}
