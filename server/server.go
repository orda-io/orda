package main

const (
	port = ":50051"
)

type server struct{}

//func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
//	log.Printf("Received: %v", in.Name)
//	return &pb.HelloReply{Message: "Hello" + in.Name}, nil
//}
//
//func main() {
//	lis, err := net.Listen("tcp", port)
//	if err != nil {
//		log.Fatalf("failed to listen: %v", err)
//	}
//	s := grpc.NewServer()
//	pb.RegisterGreeterServer(s, &server{})
//	if err := s.Serve(lis); err != nil {
//		log.Fatalf("failed to serve: %v", err)
//	}
//}
