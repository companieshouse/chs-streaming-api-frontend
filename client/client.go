package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	//"sync"
	"time"
)

type Client struct {
	baseurl string
	broker  Publishable
	client  Gettable
	wg      *sync.WaitGroup
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

func NewClient(baseurl string, broker Publishable, client Gettable) *Client {
	return &Client{
		baseurl,
		broker,
		client,
		nil,
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
		line, err := reader.ReadBytes('\n') //json object with data(consumed) and other is offset number
		if err != nil {
			fmt.Fprintf(os.Stderr, "error during resp.Body read:%s\n", err)
			continue
		}

		result := &Result{}
		json.Unmarshal(line, result)
		c.broker.Publish(result.Data)
		if c.wg != nil {
			c.wg.Done()
		}
		time.Sleep(600)
	}

}
