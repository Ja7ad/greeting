package main

import (
	"context"
	"context/proto/greeting"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"time"
)

var client greeting.ContextServiceClient
var connection *grpc.ClientConn

func ConnectGreetingClient() {
	var err error
	connection, err = grpc.Dial(
		fmt.Sprintf("localhost:9001"),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("client successfully connected to greeting server on localhost:9001")

	client = greeting.NewContextServiceClient(connection)

}

func GetClient() greeting.ContextServiceClient {
	return client
}

func main() {
	ConnectGreetingClient()
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(9*time.Second, func() {
		fmt.Println("do cancel")
		cancel()
	})
	msg, err := GetClient().Greeting(ctx, &greeting.GreetingRequest{Name: "Javad"})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(msg)
}
