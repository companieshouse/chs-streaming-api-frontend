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
	baseurl   string
	publisher Publishable
	client    Gettable
	Wg        *sync.WaitGroup
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

func NewClient(baseurl string, publisher Publishable, getter Gettable) *Client {
	return &Client{
		baseurl,
		publisher,
		getter,
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
		c.publisher.Publish(result.Data)
		if c.Wg != nil {
			c.Wg.Done()
		}
		time.Sleep(600)
	}

}
