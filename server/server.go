package main

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/model"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	port = ":50051"
)

type server struct{}

func (s *server) ProcessPushPull(ctx context.Context, in *model.PushPullRequest) (*model.PushPullReply, error) {
	log.Printf("Received: %v", proto.MarshalTextString(in))
	return &model.PushPullReply{Id: in.Id}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	model.RegisterOrtooServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
