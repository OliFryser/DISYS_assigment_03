package main

import (
	"DISYS_assigment_03/proto"
	"context"
	"flag"
	"log"
	"net"
	"sync/atomic"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedChittyChatServer
	port            string
	messageChannels map[string]chan *proto.ServerMessage
	lamportTime     int64
}

var msgStreams []proto.ChittyChat_ClientJoinServer

var (
	port = flag.String("port", "8080", "Server port")
)

// Function for incrementing lamport time
func (server *Server) IncrementLamportTime() {
	atomic.AddInt64(&server.lamportTime, 1)
}

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
		lamportTime:     1,
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

func (server *Server) Broadcast(ctx context.Context, in *proto.ChatMessage) (*proto.ChatMessage, error) {
	log.Printf("Server received message %s, from client %s", in.Message, in.SenderId)

	// Lamport time equalizes the time of the server and the client
	if in.LamportTime > int64(server.lamportTime) {
		server.lamportTime = in.LamportTime + 1
	} else {
		server.IncrementLamportTime()
	}

	chatMessage := &proto.ChatMessage{
		Message:     in.Message,
		LamportTime: server.lamportTime,
		SenderId:    in.SenderId,
	}

	for _, channel := range server.messageChannels {
		channel <- &proto.ServerMessage{
			Message: in.SenderId + " said: " + in.Message,
		}
	}

	return chatMessage, nil
}

func (server *Server) ClientJoin(in *proto.JoinRequest, msgStream proto.ChittyChat_ClientJoinServer) error {
	log.Printf("Client %s joined the chat with lamport time %d", in.SenderId, in.LamportTime)

	if server.messageChannels[in.SenderId] == nil {
		server.messageChannels[in.SenderId] = make(chan *proto.ServerMessage, 10)
	}

	response := &proto.ServerMessage{
		Message:     "Client " + in.SenderId + " has joined the chat room",
		LamportTime: server.lamportTime,
	}

	for _, channel := range server.messageChannels {
		channel <- response
	}

	for {
		select {
		case <-msgStream.Context().Done():
			return nil
		case message := <-server.messageChannels[in.SenderId]:
			msgStream.Send(message)
		}
	}
}
