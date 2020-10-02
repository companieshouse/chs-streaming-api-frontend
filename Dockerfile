FROM golang:1.12
WORKDIR /go/src/github.com/companieshouse/chs-streaming-api-frontend
COPY . .
RUN go get -d -v ./...
RUN go build -v .
ENTRYPOINT /go/src/github.com/companieshouse/chs-streaming-api-frontend/main.go
