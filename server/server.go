package main

import (
	"DISYS_assigment_03/proto"
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedChittyChatServer
	name string
	port string
}

var serverName = flag.String("name", "default", "Senders name") //Senders name? Not Servers name?
var port = flag.String("port", "8080", "Server port")

func main() {
	flag.Parse()
	log.Printf("Server is starting\n")

	server := &Server{
		name: *serverName,
		port: *port,
	}

	go launchServer(server)

	for {

	}
}

func launchServer(server *Server) {
	grpcServer := grpc.NewServer()

	listener, err := net.Listen("tcp", ":"+server.port)

	if err != nil {
		log.Fatalf("Could not create the server %v\n", err)
	}
	log.Printf("Started server at port %s\n", server.port)

	proto.RegisterChittyChatServer(grpcServer, server)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener\n")
	}
}
