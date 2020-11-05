package client

import (
	"bufio"
	"github.com/companieshouse/chs-streaming-api-frontend/logger"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	baseurl    string
	path       string
	token      string
	broker     Publishable
	httpClient Fetchable
	wg         *sync.WaitGroup
	logger     logger.Logger
	Offset     string
	closed     bool
	closing    chan bool
	finished   chan bool
	status     chan *ResponseStatus
}

type ResponseStatus struct {
	Code int
	Err  error
}

type Publishable interface {
	Publish(msg string)
	Subscribe() (chan string, error)
	Unsubscribe(subscription chan string) error
}

type Fetchable interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

//The result of the operation.
type Result struct {
	Data   string `json:"data"`
	Offset int64  `json:"offset"`
}

func NewClient(baseurl string, path string, token string, broker Publishable, client Fetchable, logger logger.Logger) *Client {
	return &Client{
		baseurl:    baseurl,
		path:       path,
		token:      token,
		broker:     broker,
		httpClient: client,
		wg:         nil,
		logger:     logger,
		closing:    make(chan bool),
		finished:   make(chan bool),
		status:     make(chan *ResponseStatus),
	}
}

func (c *Client) Connect() *ResponseStatus {
	go c.execute()
	return <-c.status
}

func (c *Client) SetOffset(offset string) {
	c.Offset = offset
}

func (c *Client) Close() {
	c.closing <- true
	<-c.finished
}

func (c *Client) execute() {
	url := c.baseurl + c.path
	if c.Offset != "" {
		url += "?timepoint=" + c.Offset
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(c.token, "")

	resp, err := c.httpClient.Do(req)

	if err != nil {
		c.closed = true
		c.logger.Error(err)
		c.status <- &ResponseStatus{Err: err}
		return
	}
	c.status <- &ResponseStatus{Code: resp.StatusCode}
	if resp.StatusCode != http.StatusOK {
		c.closed = true
		return
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

func (c *Client) loop(reader *bufio.Reader) {
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			c.logger.Error(err)
			return
		}
		if len(line) > 1 {
			c.broker.Publish(string(line))
		}
		if c.wg != nil {
			c.wg.Done()
		}
		time.Sleep(600)
	}
}
