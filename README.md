# chs-streaming-api-frontend
This service is to publish incoming messages to all connected users.
This frontend component is outputting what it receive from chs-streaming-api-cache directly to all connected users.

#Requirements
In order to build this service you need:

Git
go

In order to run this service you need:
The service is self contained.
However, this service has a depenedency of chs-streaming-api-cache

# Getting started
Configure your service if you wish to override any of the defaults

To Run main.go

    go get
    go run main.go

To Run tests

    go build ./...
    go test ./...

format the code

    go fmt ./...

If test failing to run
    export GO111MODULE="on"

# Dockerise build and deployment of service

Refer PR - https://github.com/companieshouse/docker-chs-development/pull/59


# Environment Variables

Name                   | Description                          | Mandatory | Default | Example
---------------------- | ---------------------------------------------------------- | --------- | ------- | ------------------
CACHE_BROKER_URL       | The host URL of the CACHE service    | ✓         |         | http://HOST:PORT
HEARTBEAT_INTERVAL     | Heart Beat Interval                  | ✓         |         | 30
REQUEST_TIMEOUT        | Request Timeout                      | ✓         |         | 86400