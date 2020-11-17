# chs-streaming-api-frontend

## Contents

The Companies House Streaming Platform Frontend consumes offsets from the Streaming Platform Cache service and pushes these to connected users as an event stream.

## Requirements

The following services and applications are required to build and/or run chs-streaming-api-frontend:

* AWS ECR
* Docker
* Redis
* Companies House Streaming Platform Cache (chs-streaming-api-cache)

You will need an HTTP client that supports server-sent events (e.g. cURL) to connect to the service and receive published offsets.

## Building and Running Locally

1. Login to AWS ECR.
2. Build the project by running `docker build` from the project directory.
3. Run the Docker image that has been built by running `docker run IMAGE_ID` from the command line, ensuring values have been specified for the environment variables (see Configuration) and that port 6002 is exposed.
4. Send a GET request using your HTTP client to /filings. A connection should be established and any offsets pushed by chs-streaming-api-cache should appear in the response body.
5. To begin consuming offsets from a particular timepoint, set the timepoint query parameter in the request URL to a valid offset number.

## Configuration

Variable|Description|Example|Mandatory|
--------|-----------|-------|---------|
CACHE_BROKER_URL|The URL of the Streaming API Cache service|http://chs-streaming-api-cache:6001|yes
REQUEST_TIMEOUT|The duration that a user can remain connected to the service|1200|yes
HEARTBEAT_INTERVAL|The interval at which newlines will be echoed back to a connected user to verify connectivity|30|yes
