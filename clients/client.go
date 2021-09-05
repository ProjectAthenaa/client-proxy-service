package clients

import (
	"context"
	client_proxy "github.com/ProjectAthenaa/sonic-core/protos/clientProxy"
	"github.com/prometheus/common/log"
	"strings"
)

//registerNewClient creates a new client and returns a pointer to it
func registerNewClient(stream client_proxy.Proxy_RegisterServer) *client {
	return &client{
		Proxy_RegisterServer: stream,
		requests:             make(chan *client_proxy.Request),
		responses:            make(map[string]chan *client_proxy.Response),
	}
}

type client struct {
	client_proxy.Proxy_RegisterServer
	requests  chan *client_proxy.Request
	responses map[string]chan *client_proxy.Response
}

//process listens to responses from all of the localhost proxies and forwards them to the appropriate channels as well as
//listening to new requests and proxying them to the localhost proxies
func (c *client) process(ctx context.Context) error {
	go func() {
		var req *client_proxy.Request
		for {
			select {
			case req = <-c.requests:
				if err := c.Send(req); err != nil {
					log.Error("err sending req ", err)
					continue
				}
			case <-ctx.Done():
				return
			default:
				break
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.Recv()
			if err != nil {
				if strings.Contains(err.Error(), "context canceled") {
					continue
				}
				log.Error("err receiving req", err.Error())
				return err
			}

			if respChan, ok := c.responses[msg.TaskID]; ok {
				respChan <- msg
				delete(c.responses, msg.TaskID)
			}
		}
	}
}

//doRequest creates a request wish and returns a channel to listen for responses
func (c *client) doRequest(req *client_proxy.Request) <-chan *client_proxy.Response {
	var respChan = make(chan *client_proxy.Response)
	c.responses[req.TaskID] = respChan
	c.requests <- req
	return respChan
}
