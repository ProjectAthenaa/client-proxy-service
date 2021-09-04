package clients

import (
	"context"
	client_proxy "github.com/ProjectAthenaa/sonic-core/protos/clientProxy"
	"github.com/prometheus/common/log"
)

func RegisterNewClient(stream client_proxy.Proxy_RegisterServer) *Client {
	return &Client{
		Proxy_RegisterServer: stream,
		requests:             make(chan *client_proxy.Request),
		responses:            make(map[string]chan *client_proxy.Response),
	}
}

type Client struct {
	client_proxy.Proxy_RegisterServer
	requests  chan *client_proxy.Request
	responses map[string]chan *client_proxy.Response
}

func (c *Client) Process(ctx context.Context) error {
	go func() {
		for req := range c.requests {
			if err := c.Send(req); err != nil {
				log.Error("err sending req ", err)
				continue
			}
		}
	}()

	for {
		msg, err := c.Recv()
		if err != nil {
			log.Error("err reciving req")
			return err
		}

		if respChan, ok := c.responses[msg.TaskID]; ok {
			respChan <- msg
			delete(c.responses, msg.TaskID)
		}
	}
}

func (c *Client) doRequest(req *client_proxy.Request) <- chan *client_proxy.Response {
	var respChan = make(chan *client_proxy.Response)
	c.responses[req.TaskID] = respChan
	c.requests <- req
	return respChan
}
