package main

import (
	"fmt"
	"io/ioutil"
	"context"
	"net/http"
	"google.golang.org/grpc"
	pb "otdd.io/example/dependent-grpc/grpc"
	"github.com/apache/thrift/lib/go/thrift"
	service "otdd.io/example/dependent-thrift/service"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (p *Handler) SayHello(ctx context.Context, req string) (resp string, err error) {
	fmt.Print("received req:",req)
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
	return ret, nil
}
