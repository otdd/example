package main
 
import (
	"io"
	"io/ioutil"
	"log"
	"fmt"
	"strings"
	"context"
	"net/http"
	"google.golang.org/grpc"
	pb "otdd.io/example/dependent-grpc/grpc"
	"github.com/apache/thrift/lib/go/thrift"
	service "otdd.io/example/dependent-thrift/service"
)
 
func SayHello(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received: %s", formatRequest(r))
	// call http service
	httpResp, err := http.Get("http://service-http:8080/SayHello")
	if err != nil {
		io.WriteString(w, "Failed to call service-http:"+err.Error()+"\n")
	} else {
		defer httpResp.Body.Close()
		body, err := ioutil.ReadAll(httpResp.Body)
		if err != nil {
			io.WriteString(w, "Failed to read service-http resp:"+err.Error()+"\n")
		} else {
			io.WriteString(w, "Resp from service-http:"+string(body)+"\n")
		}
	}
	// call grpc service
	conn, err := grpc.Dial("service-grpc:8764",grpc.WithInsecure())
	if err != nil {
	    io.WriteString(w, "Failed to connect service-grpc:"+err.Error()+"\n")
	} else {
		defer conn.Close()
		client := pb.NewSayHelloServiceClient(conn)
		grpcResp, err := client.SayHello(context.Background(), &pb.SayHelloReq{Req: "Hello OTDD"})
		if err != nil {
			io.WriteString(w, "Failed to call service-grpc :"+err.Error()+"\n")
		} else {
			io.WriteString(w, "Resp from service-grpc:"+grpcResp.Resp+"\n")
		}
	}
	// call thrift service
	trans, err := thrift.NewTSocket("service-thrift:9001")
	if err != nil {
		io.WriteString(w, "Failed to connect service-thrift:"+err.Error()+"\n")
	} else {
		transportFactory := thrift.NewTTransportFactory()
		transport, _ := transportFactory.GetTransport(trans)
		defer transport.Close()
		if err := transport.Open(); err != nil {
			io.WriteString(w, "Failed to connect service-thrift:"+err.Error()+"\n")
		} else {
			protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
			iprot := protocolFactory.GetProtocol(transport)
			oprot := protocolFactory.GetProtocol(transport)
			client := service.NewSayHelloServiceClient(thrift.NewTStandardClient(iprot, oprot))
			defaultCtx := context.Background()
			thriftResp, err := client.SayHello(defaultCtx, "Hello OTDD")
			if err != nil {
				io.WriteString(w, "Failed to call service-thrift:"+err.Error()+"\n")
			} else {
				io.WriteString(w, "Resp from service-thrift:"+thriftResp+"\n")
			}
		}
	}
}

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}
	
	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	} 
	 // Return the request as a string
	 return strings.Join(request, "\n")
}
 
func main() {
	http.HandleFunc("/SayHello", SayHello)
	http.ListenAndServe(":8080", nil)
}
