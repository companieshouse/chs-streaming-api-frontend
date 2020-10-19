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
	baseurl   string
	publisher Publishable
	client    Gettable
	Wg        *sync.WaitGroup
	logger    logger.Logger
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

func NewClient(baseurl string, publisher Publishable, getter Gettable, logger logger.Logger) *Client {
	return &Client{
		baseurl,
		publisher,
		getter,
		nil,
		logger,
	}
}

func (c *Client) Connect() {
	resp, _ := c.client.Get(c.baseurl)
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

		c.publisher.Publish(string(line))
		if c.Wg != nil {
			c.Wg.Done()
		}
		time.Sleep(600)
	}

}
