package clients

import (
	"context"
	"errors"
	client_proxy "github.com/ProjectAthenaa/sonic-core/protos/clientProxy"
	"github.com/google/uuid"
	"github.com/prometheus/common/log"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	client_proxy.UnimplementedProxyServer
	clients map[string]*client
}

func NewServer() *Server {
	return &Server{
		clients: make(map[string]*client),
	}
}

//Do proxies the request received by an internal service to the appropriate client
func (s *Server) Do(ctx context.Context, request *client_proxy.Request) (*client_proxy.Response, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no_user_id")
	}

	userIDArr := md.Get("UserID")
	if len(userIDArr) == 0 {
		return nil, errors.New("no_user_id")
	}

	userID := userIDArr[0]

	//create unique task id to match request with response
	request.TaskID = uuid.NewString() + request.URL
	log.Info("New Request: ", request.TaskID)
	//check if client is connected
	_, ok = s.clients[userID]
	if !ok {
		return nil, errors.New("client_not_connected")
	}

	//do request with the connected client
	responses := s.clients[userID].doRequest(request)

	var resp *client_proxy.Response

	for {
		select {
		//listen to responses from the client
		case resp = <-responses:
			log.Info("New Response: ", resp.TaskID)
			log.Info("Headers: ", resp.Headers)
			if e, ok := resp.Headers["ERROR"]; ok {
				log.Info(e)
				return nil, errors.New(e)
			}
			return resp, nil
		case <-ctx.Done():
			return nil, errors.New("timeout")
		default:
			continue
		}
	}
}

//Register registers a user's localhost connection to the proxy service
func (s *Server) Register(stream client_proxy.Proxy_RegisterServer) error {
	//retrieve data from metadata
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		//if metadata not existent stop the proxy as it cannot be authenticated
		return stream.Send(&client_proxy.Request{Headers: map[string]string{"STOP": "1"}})
	}
	//get user id and check if its valid
	userIDArr := md.Get("UserID")
	if len(userIDArr) == 0 {
		return stream.Send(&client_proxy.Request{Headers: map[string]string{"STOP": "1"}})
	}

	_ = stream.Send(&client_proxy.Request{Headers: map[string]string{"STOP": "0"}})

	userID := userIDArr[0]

	//instantiate client
	c := registerNewClient(stream)

	//add client to available clients in the server
	s.clients[userID] = c

	//delete the client from clients after it's done processing
	defer delete(s.clients, userID)

	//start processing requests/responses
	return c.process(stream.Context())
}
