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
	port            string
	messageChannels map[string]chan *proto.ServerMessage
}

var msgStreams []proto.ChittyChat_ClientJoinServer

var (
	port = flag.String("port", "8080", "Server port")
)

func main() {
	// Prints to log file instead of terminal
	/* f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f) */

	flag.Parse()
	log.Printf("Server is starting\n")

	server := &Server{
		port:            *port,
		messageChannels: make(map[string]chan *proto.ServerMessage),
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

	for _, channel := range s.messageChannels {
		channel <- &proto.ServerMessage{
			Message:     in.SenderId + " said: " + in.Message,
			LamportTime: in.LamportTime + 1,
		}
	}

	return chatMessage, nil
}

func (s *Server) ClientJoin(in *proto.JoinRequest, msgStream proto.ChittyChat_ClientJoinServer) error {
	log.Printf("Client %s joined the chat with lamport time %d", in.SenderId, in.LamportTime)

	if s.messageChannels[in.SenderId] == nil {
		log.Printf("Making channel for client %s\n", in.SenderId)
		s.messageChannels[in.SenderId] = make(chan *proto.ServerMessage)
	}
	log.Printf("Length of the message channels: %d\n", len(s.messageChannels))

	response := &proto.ServerMessage{
		Message:     "Client " + in.SenderId + " has joined the chat room",
		LamportTime: int64(1),
	}

	for _, channel := range s.messageChannels {
		log.Printf("Entered the loop\n")
		channel <- response
		log.Print("Put the message in channel\n")
	}

	for {
		select {
		case <-msgStream.Context().Done():
			log.Printf("Client Left")
			return nil
		case message := <-s.messageChannels[in.SenderId]:
			log.Printf("Sending message\n")
			msgStream.Send(message)
		}
	}
}
