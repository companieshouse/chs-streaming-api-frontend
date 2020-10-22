package client

import (
	"bufio"
	"fmt"
	"github.com/companieshouse/chs-streaming-api-frontend/logger"
	"github.com/companieshouse/chs.go/log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Client struct {
	baseurl    string
	path       string
	broker     Publishable
	httpClient Gettable
	Wg         *sync.WaitGroup
	logger     logger.Logger
}

type Publishable interface {
	Publish(msg string)
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
		baseurl,
		path,
		broker,
		client,
		nil,
		logger,
	}
}

func (c *Client) Connect() {
	resp, err := c.httpClient.Get(c.baseurl + c.path)

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
}

func (c *Client) loop(reader *bufio.Reader) {

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			c.logger.Error(err, log.Data{})
			fmt.Fprintf(os.Stderr, "error during resp.Body read:%s\n", err)
			continue
		}

		c.broker.Publish(string(line))
		if c.Wg != nil {
			c.Wg.Done()
		}
		time.Sleep(600)
	}

}
