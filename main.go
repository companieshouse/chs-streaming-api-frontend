package main

import (
	"fmt"
	"github.com/companieshouse/chs-streaming-api-frontend/config"
	"github.com/companieshouse/chs-streaming-api-frontend/handlers"
	"github.com/companieshouse/chs-streaming-api-frontend/logger"
	"time"

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
	}

	streamHandler.AddStream(svc.Router(), "/filings", filingHistoryStream, cfg.CacheBrokerURL)
	streamHandler.AddStream(svc.Router(), "/companies", companyProfileStream, cfg.CacheBrokerURL)
	streamHandler.AddStream(svc.Router(), "/insolvency-cases", companyInsolvencyStream, cfg.CacheBrokerURL)
	streamHandler.AddStream(svc.Router(), "/charges", companyChargesStream, cfg.CacheBrokerURL)
	streamHandler.AddStream(svc.Router(), "/officers", companyOfficersStream, cfg.CacheBrokerURL)
	streamHandler.AddStream(svc.Router(), "/persons-with-significant-control", companyPSCStream, cfg.CacheBrokerURL)

	svc.Start()
}
