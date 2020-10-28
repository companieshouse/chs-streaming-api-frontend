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

# Docker build

- Reconfigure your AWS account to use eu-west-1 by running 

        "aws configure"
   This is to making sure you accept default values for access key and secret key

- Login to ECR using your AWS account by running 

        docker login -u AWS -p "$(aws ecr get-login-password)" https://169942020521.dkr.ecr.eu-west-1.amazonaws.com
        
        
  NOTE- This is the container registry. 
        This speeds up with fetching services images rather than cloning and building the services at the local workstation.

- Build the project by running


     docker run <IMAGE_ID>
        
     IMAGE_ID is 169942020521.dkr.ecr.eu-west-1.amazonaws.com/local/chs-streaming-api-frontend:latest



- Send a GET request using your HTTP client to /filings.
  A connection should be established and ready to receive the messaage from backend and cache service

        $ curl -v http://stream.chs.local/filings
        
        $ curl -v http://stream.chs.local/filings?timepoint=1

Note - Other than Environment variables, ensure that value, in chs-config, have been specified for the 
       KAFKA_STREAMING_BROKER_ADDR as kafka2.broker:29095
       This is needed as kafka streaming is running on port 29095

# Environment Variables

Name                   | Description                          | Mandatory | Default | Example
---------------------- | ---------------------------------------------------------- | --------- | ------- | ------------------
CACHE_BROKER_URL       | The host URL of the CACHE service    | ✓         |         | http://HOST:PORT
HEARTBEAT_INTERVAL     | Heart Beat Interval                  | ✓         |         | 30
REQUEST_TIMEOUT        | Request Timeout                      | ✓         |         | 86400