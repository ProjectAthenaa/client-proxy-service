package main

import (
	"client-proxy-service/clients"
	client_proxy "github.com/ProjectAthenaa/sonic-core/protos/clientProxy"
	"github.com/prometheus/common/log"
	"google.golang.org/grpc"
	"net"
	"os"
)

func main() {
	var lis net.Listener
	var err error
	if os.Getenv("DEBUG") == "1" {
		lis, err = net.Listen("tcp", ":8080")
	} else {
		lis, err = net.Listen("tcp", ":3000")
	}
	if err != nil {
		log.Fatalln(err)
	}

	server := grpc.NewServer()

	client_proxy.RegisterProxyServer(server, clients.NewServer())
	log.Info("Started proxy on localhost:8080")
	if err = server.Serve(lis); err != nil {
		log.Fatalln(err)
	}

}
