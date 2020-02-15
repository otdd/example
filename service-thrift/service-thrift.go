package main
 
import (
	"github.com/apache/thrift/lib/go/thrift"
	service "otdd.io/example/dependent-thrift/service"
)
 
func main() {
	addr := "0.0.0.0:9001"
        protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
        transportFactory := thrift.NewTTransportFactory()
        transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
        transport, _ := thrift.NewTServerSocket(addr)
        handler := NewHandler()
        processor := service.NewSayHelloServiceProcessor(handler)
        server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
        server.Serve()
}
