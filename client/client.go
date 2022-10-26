package main

import (
	"DISYS_assigment_03/proto"
	"bufio"
	"context"
	"flag"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	id   int
	port string
}

var (
	clientPort = flag.String("cPort", "8081", "client port number")
	serverPort = flag.String("sPort", "8080", "server port number (should match the port used for the server")
)

func main() {
	flag.Parse()

	client := &Client{
		id:   1,
		port: *clientPort,
	}

	go WaitForChatMessage(client)

	for {

	}
}

func WaitForChatMessage(client *Client) {
	serverConnection, _ := connectToServer()

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := scanner.Text()
		log.Printf("Client typed chat message: %s\n", input)

		chatMessage, err := serverConnection.Broadcast(context.Background(), &proto.ChatMessage{
			Message:     input,
			LamportTime: int64(1),
			SenderId:    int64(client.id),
		})
		if err != nil {
			log.Fatalf("Failed to send chatmessage\n")
		}

		log.Printf("Client recieved message: \"%s\" from server\n", chatMessage.Message)
	}

}

func connectToServer() (proto.ChittyChatClient, error) {
	conn, err := grpc.Dial("localhost:"+*serverPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %s\n", *serverPort)
	}
	log.Printf("Connected to server port %s\n", *serverPort)
	return proto.NewChittyChatClient(conn), nil
}
