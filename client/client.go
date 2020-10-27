package client

import (
	"bufio"
	"github.com/companieshouse/chs-streaming-api-frontend/logger"
	"github.com/companieshouse/chs.go/log"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	baseurl    string
	path       string
	broker     Publishable
	httpClient Gettable
	wg         *sync.WaitGroup
	logger     logger.Logger
	Offset     string
	closed     bool
	closing    chan bool
	finished   chan bool
}

type Publishable interface {
	Publish(msg string)
	Subscribe() (chan string, error)
	Unsubscribe(subscription chan string) error
}

type Gettable interface {
	Get(url string) (resp *http.Response, err error)
}

//The result of the operation.
type Result struct {
	Data   string `json:"data"`
	Offset int64  `json:"offset"`
}

func NewClient(baseurl string, path string, broker Publishable, client Gettable, logger logger.Logger) *Client {
	return &Client{
		baseurl:    baseurl,
		path:       path,
		broker:     broker,
		httpClient: client,
		wg:         nil,
		logger:     logger,
		closing:    make(chan bool),
		finished:   make(chan bool),
	}
}

func (c *Client) Connect() {
	url := c.baseurl + c.path
	if c.Offset != "" {
		url += "?timepoint=" + c.Offset
	}
	resp, err := c.httpClient.Get(url)

	if err != nil {
		c.logger.Error(err, log.Data{})
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		c.logger.Info("Unable to connect to cache broker from endpoint", log.Data{"endpoint": c.baseurl, "Http Status": resp.StatusCode})
		panic("Unable to connect to cache broker from endpoint")
	}

	body := resp.Body
	reader := bufio.NewReader(body)
	go c.loop(reader)
	<-c.closing
	err = body.Close()
	if err != nil {
		c.logger.Error(err)
	}
	c.closed = true
	c.finished <- true
}

func (c *Client) SetOffset(offset string) {
	c.Offset = offset
}

func (c *Client) loop(reader *bufio.Reader) {

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			c.logger.Error(err)
			return
		}

		c.broker.Publish(string(line))
		if c.wg != nil {
			c.wg.Done()
		}
		time.Sleep(600)
	}

}

func (c *Client) Close() {
	c.closing <- true
	<-c.finished
}
