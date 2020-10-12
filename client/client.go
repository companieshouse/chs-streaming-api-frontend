package client

import (
	"net/http"
)

type Client struct {
	baseurl string
	broker  Publishable
	client  Gettable
}

type Publishable interface {
	Publish (msg string)
}

type Gettable interface {
	Get(url string) (resp *http.Response, err error)
}

func NewClient(baseurl string, broker Publishable, client Gettable) *Client{
	return &Client {
		baseurl,
		broker,
		client,
	}
}

