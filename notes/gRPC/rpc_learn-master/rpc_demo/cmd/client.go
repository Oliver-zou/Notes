package main

import (
	"context"
	"fmt"
	"github.com/ldaysjun/rpc_learn/protobuf/helloworld"
	"github.com/ldaysjun/rpc_learn/rpc_demo/pkg/api"
	"log"
	"time"
)

func main() {
	client := api.NewGreeterClient()
	for i := 0;i<1 ;i++  {
		fmt.Println("1")
		req := &helloworld.HelloRequest{
			Name: "ldaysjun",
		}
		ctx, _ := context.WithTimeout(context.Background(), time.Second * 3)
		rsp, err := client.SayHello(ctx, req)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		fmt.Println("rsp.message = ", rsp.Message)
	}
	time.Sleep(time.Hour)
}


