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
	"github.com/go-redis/redis"
	"database/sql"
	_"github.com/go-sql-driver/mysql"
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
	// call redis dependent
	client := redis.NewClient(&redis.Options{ 
		Addr: "dependent-redis:6379",
	})
	err = client.Set("key", "Hello OTDD", 0).Err()
	if err != nil {
		ret = ret + "Failed to set key to dependent-redis:"+err.Error()+"\n"
	} else {
		val, err := client.Get("key").Result() 
		if err != nil { 
			ret = ret + "Failed to get value from dependent-redis:"+err.Error()+"\n"
		} else {
			ret = ret + "value from dependent-redis:"+val+"\n"
		}
	}

	// call mysql dependent
	db, err := sql.Open("mysql","root:123456@tcp(dependent-mysql:3306)/hello_otdd")
	if err != nil {
		ret = ret + "Failed to connect to dependent-mysql:"+err.Error()+"\n"
	} else {
		defer db.Close()
		stmt, err := db.Prepare("insert into hello_otdd_tb(value) VALUES(?)")
		if err != nil {
			ret = ret + "failed to prepare insert statement:"+err.Error()+"\n"
		}
		res, err := stmt.Exec("Hello OTDD")
		if err != nil {
			ret = ret + "failed to execute insert statement:"+err.Error()+"\n"
		} else {
			id, _ := res.LastInsertId()
			ret = ret + fmt.Sprintf("inserted row to table, id:%d \n",id)
			rows, err := db.Query("select * from hello_otdd_tb order by id desc limit 1")
			if err != nil {
				ret = ret + "Failed to select from dependent-mysql:"+err.Error()+"\n"
			} else {
				//https://kylewbanks.com/blog/query-result-to-map-in-golang
				cols, _ := rows.Columns()
				for rows.Next() {
					columns := make([]interface{}, len(cols))
					columnPointers := make([]interface{}, len(cols))
					for i, _ := range columns {
						columnPointers[i] = &columns[i]
					}
					if err := rows.Scan(columnPointers...); err != nil {
						ret = ret + fmt.Sprintf("failed to parse row from dependent-mysql:%v \n", rows)
					} else {
						 m := make(map[string]interface{})
						for i, colName := range cols {
							val := columnPointers[i].(*interface{})
							m[colName] = *val
						}
						ret = ret + fmt.Sprintf("row with max id from dependent-mysql:%v \n", m)
					}
				}
			}
		}
	}
	return ret, nil
}
