package clients

import (
	"context"
	"errors"
	client_proxy "github.com/ProjectAthenaa/sonic-core/protos/clientProxy"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	client_proxy.UnimplementedProxyServer
	clients map[string]*Client
}

func NewServer() *Server {
	return &Server{
		clients: make(map[string]*Client),
	}
}

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

	responses := s.clients[userID].doRequest(request)

	var resp *client_proxy.Response

	for {
		select {
		case resp = <-responses:
			return resp, nil
		case <-ctx.Done():
			return nil, errors.New("timeout")
		default:
			continue
		}
	}
}

func (s *Server) Register(stream client_proxy.Proxy_RegisterServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return stream.Send(&client_proxy.Request{Headers: map[string]string{"STOP": ""}})
	}

	userIDArr := md.Get("UserID")
	if len(userIDArr) == 0 {
		return stream.Send(&client_proxy.Request{Headers: map[string]string{"STOP": ""}})
	}

	userID := userIDArr[0]

	client := RegisterNewClient(stream)

	s.clients[userID] = client

	return client.Process(stream.Context())
}
