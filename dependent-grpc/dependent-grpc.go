package main

import (
        "context"
        "log"
        "net"
        "google.golang.org/grpc"
        pb "otdd.io/example/dependent-grpc/grpc"
)

const (
        port = ":8764"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
        pb.UnimplementedSayHelloServiceServer
}

func (s *server) SayHello(ctx context.Context, in *pb.SayHelloReq) (*pb.SayHelloResp, error) {
        log.Printf("Received: %v", in.GetReq())
        return &pb.SayHelloResp{Resp: "Hello OTDD"}, nil
}

func main() {
        lis, err := net.Listen("tcp", port)
        if err != nil {
                log.Fatalf("failed to listen: %v", err)
        }
        s := grpc.NewServer()
        pb.RegisterSayHelloServiceServer(s, &server{})
        if err := s.Serve(lis); err != nil {
                log.Fatalf("failed to serve: %v", err)
        }
}
