package main

import (
	"github.com/ldaysjun/rpc_learn/protobuf/helloworld"
	"github.com/ldaysjun/rpc_learn/rpc_demo/internal/logic"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const addr = "localhost:50052"

func main() {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()
	imp, err := logic.NewGreeter()
	if err != nil {
		panic(err)
	}
	// 注册服务
	imp.ServiceRegister(logic.Scheme,logic.ServiceName, addr)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		if i, ok := s.(syscall.Signal); ok {
			// 退出注销服务
			imp.ServiceLogout(logic.Scheme,logic.ServiceName,addr)
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}
	}()

	helloworld.RegisterGreeterServer(s, imp)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
