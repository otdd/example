package main


import (
	"fmt"
	"context"
//	"otdd.io/example/dependent-thrift/service"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (p *Handler) SayHello(ctx context.Context, req string) (resp string, err error) {
	fmt.Print("received req:",req)
	return "Hello OTDD" , nil
}
