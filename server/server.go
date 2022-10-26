package main

import (
	"DISYS_assigment_03/proto"
	"context"
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedChittyChatServer
	port string
}

var port = flag.String("port", "8080", "Server port")

func main() {
	flag.Parse()
	log.Printf("Server is starting\n")

	server := &Server{
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

func (s *Server) Broadcast(ctx context.Context, in *proto.ChatMessage) (*proto.ChatMessage, error) {
	log.Printf("Server received message %s, from client %s", in.Message, in.SenderId)

	chatMessage := &proto.ChatMessage{
		Message:     in.Message,
		LamportTime: int64(1),
		SenderId:    in.SenderId,
	}

	return chatMessage, nil

}
