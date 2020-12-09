package api

import (
	"context"
	"github.com/ldaysjun/rpc_learn/protobuf/helloworld"
	"github.com/ldaysjun/rpc_learn/rpc_demo/internal/utils/balancer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"log"
	"time"
)

func NewGreeterClient() helloworld.GreeterClient {
	r := &balancer.ETCDResolverBuilder{}
	resolver.Register(r)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	conn, err := grpc.DialContext(ctx, r.Scheme()+"://rpc/demo.hello.world", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := helloworld.NewGreeterClient(conn)
	return c
}
