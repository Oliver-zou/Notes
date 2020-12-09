package logic

import (
	"errors"
	"fmt"
	"github.com/ldaysjun/rpc_learn/protobuf/helloworld"
	"golang.org/x/net/context"
)

func (g *greeter) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	// 检查当前是否允许该事件通过
	if !g.limiter.Allow(){
		return nil,errors.New("service is busy, please try again later")
	}
	rsp := &helloworld.HelloReply{
		Message: fmt.Sprintf("hello:%s", req.Name),
	}
	return rsp, nil
}
