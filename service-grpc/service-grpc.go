package main
 
import (
	"io/ioutil"
	"log"
	"context"
	"net"
	"net/http"
	"google.golang.org/grpc"
	pb "otdd.io/example/dependent-grpc/grpc"
	"github.com/apache/thrift/lib/go/thrift"
	service "otdd.io/example/dependent-thrift/service"
)

const (
        port = ":8764"
)

type server struct {
        pb.UnimplementedSayHelloServiceServer
}

func (s *server) SayHello(ctx context.Context, in *pb.SayHelloReq) (*pb.SayHelloResp, error) {
	log.Printf("Received: %v", in.GetReq())
	ret := ""
	// call http dependent
	httpResp, err := http.Get("http://dependent-http:8080/SayHello")
	if err != nil {
		ret = ret + "Failed to call dependent-http:"+err.Error()+"\n"
	} else {
		defer httpResp.Body.Close()
		body, err := ioutil.ReadAll(httpResp.Body)
		if err != nil {
			ret = ret + "Failed to read dependent-http resp:"+err.Error()+"\n"
		} else {
			ret = ret + "Resp from dependent-http:"+string(body)+"\n"
		}
	}
	// call grpc dependent
	conn, err := grpc.Dial("dependent-grpc:8764",grpc.WithInsecure())
	if err != nil {
	    ret = ret + "Failed to connect dependent-grpc:"+err.Error()+"\n"
	} else {
		defer conn.Close()
		client := pb.NewSayHelloServiceClient(conn)
		grpcResp, err := client.SayHello(context.Background(), &pb.SayHelloReq{Req: "Hello OTDD"})
		if err != nil {
			ret = ret + "Failed to call dependent-grpc :"+err.Error()+"\n"
		} else {
			ret = ret + "Resp from dependent-grpc:"+grpcResp.Resp+"\n"
		}
	}
	// call thrift dependent
	trans, err := thrift.NewTSocket("dependent-thrift:9001")
	if err != nil {
		ret = ret + "Failed to connect dependent-thrift:"+err.Error()+"\n"
	} else {
		transportFactory := thrift.NewTTransportFactory()
		transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
		transport, _ := transportFactory.GetTransport(trans)
		defer transport.Close()
		if err := transport.Open(); err != nil {
			ret = ret + "Failed to connect dependent-thrift:"+err.Error()+"\n"
		} else {
			protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
			iprot := protocolFactory.GetProtocol(transport)
			oprot := protocolFactory.GetProtocol(transport)
			client := service.NewSayHelloServiceClient(thrift.NewTStandardClient(iprot, oprot))
			defaultCtx := context.Background()
			thriftResp, err := client.SayHello(defaultCtx, "Hello OTDD")
			if err != nil {
				ret = ret + "Failed to call dependent-thrift:"+err.Error()+"\n"
			} else {
				ret = ret +  "Resp from dependent-thrift:"+thriftResp+"\n"
			}
		}
	}
	return &pb.SayHelloResp{Resp: ret}, nil
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
