package main

import (
	"context"
	"context/proto/greeting"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"time"
)

type Server struct {
	greeting.UnimplementedContextServiceServer
}

func timeoutInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var err error
	var result interface{}
	done := make(chan struct{})
	go func() {
		result, err = handler(ctx, req)
		done <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			msg := "Client cancelled, abandoning."
			log.Println(msg)
			return nil, status.New(codes.Canceled, msg).Err()
		}
	case <-done:
	}
	return result, err
}

func (Server) Greeting(ctx context.Context, req *greeting.GreetingRequest) (*greeting.GreetingResponse, error) {
	msg, err := greetingManager(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	return &greeting.GreetingResponse{
		Message: msg,
	}, nil
}

func greetingManager(ctx context.Context, name string) (string, error) {
	result := make(chan string)
	go greetingMessage(ctx, name, result)
	for {
		select {
		case <-ctx.Done():
			return "", errors.New("context cancel")
		case r := <-result:
			return r, nil
		}
	}
}

func greetingMessage(ctx context.Context, name string, result chan string) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("request for do greeting by " + name)
			result <- fmt.Sprintf("hello %s", name)
		}
	}
}

func main() {
	listen, err := net.Listen("tcp", "localhost:9001")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("ran greeting service on localhost:9001")
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(timeoutInterceptor))
	greeting.RegisterContextServiceServer(srv, &Server{})
	if err := srv.Serve(listen); err != nil {
		log.Fatalln(err)
	}
}
