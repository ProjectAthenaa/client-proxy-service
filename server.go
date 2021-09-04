package main

import (
	"client-proxy-service/clients"
	client_proxy "github.com/ProjectAthenaa/sonic-core/protos/clientProxy"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalln(err)
	}

	server := grpc.NewServer()

	client_proxy.RegisterProxyServer(server, clients.NewServer())

	if err = server.Serve(lis); err != nil {
		log.Fatalln(err)
	}

}
