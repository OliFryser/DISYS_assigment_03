package main

import (
	"DISYS_assigment_03/proto"
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	name        string
	port        string
	lamportTime int64
}

var (
	clientPort = flag.String("cPort", "8081", "client port number")
	serverPort = flag.String("sPort", "8080", "server port number (should match the port used for the server)")
	clientName = flag.String("name", "unknown", "name of the client")
	joined     bool
)

// Function for incrementing lamport time
func (client *Client) IncrementLamportTime() {
	atomic.AddInt64(&client.lamportTime, 1)
}

func main() {
	flag.Parse()

	client := &Client{
		name:        *clientName,
		port:        *clientPort,
		lamportTime: 1,
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
		// If the user attempts to send a message longer than 128 characters, the message gets rejected
		if len(input) > 128 {
			log.Println("Message is too long, please try again.")
			continue
		}

		if input == "/join" && !joined {
			messageStream, err := serverConnection.ClientJoin(context.Background(), &proto.JoinRequest{
				LamportTime: client.lamportTime,
				SenderId:    client.name,
			})

			if err != nil {
				log.Fatalf("Failed to join the chatroom.\n")
			}

			joined = true
			/*
				message, err2 := messageStream.Recv()
				if err2 != nil {
					log.Fatalf("Failed to recieve message.\n")
				}
				log.Printf("%s (lamport time: %d)", message.Message, message.LamportTime)
			*/
			go ReceiveMessages(messageStream, *client)
			continue
		}
		if !joined {
			log.Printf("You have to join the chatroom first. Type \"/join\" to join the chatroom.")
			continue
		}
		_, err := serverConnection.Broadcast(context.Background(), &proto.ChatMessage{
			Message:     input,
			LamportTime: client.lamportTime,
			SenderId:    client.name,
		})
		if err != nil {
			log.Fatalf("Failed to send chatmessage %s\n", err.Error())
		}
		//Increments local time when sending a message
		client.IncrementLamportTime()
		//log.Printf("Client recieved message: \"%s\" from server\n", chatMessage.Message)
	}

}

func ReceiveMessages(messageStream proto.ChittyChat_ClientJoinClient, client Client) {
	//done := make(chan bool)
	//go func() {
	for joined {
		message, err := messageStream.Recv()
		if err == io.EOF {
			log.Printf("Done")
			//done <- true
			return
		}
		if err != nil {
			log.Fatalf("Failed to receive message")
		}

		//Picks highest value lamport time and increments it
		if message.LamportTime > int64(client.lamportTime) {
			client.lamportTime = message.LamportTime + 1
		} else {
			client.IncrementLamportTime()
		}

		log.Printf("%s (lamport time: %d)", message.Message, client.lamportTime)
	}
	//}()
	//<-done
}

func connectToServer() (proto.ChittyChatClient, error) {
	conn, err := grpc.Dial("localhost:"+*serverPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %s\n", *serverPort)
	}
	log.Printf("Connected to server port %s\n", *serverPort)
	return proto.NewChittyChatClient(conn), nil
}
