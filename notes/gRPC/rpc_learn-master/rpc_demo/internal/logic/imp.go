package logic

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"golang.org/x/time/rate"
	"log"
	"time"
)

const (
	ServiceName = "demo.hello.world"
	Scheme      = "test"
)

type greeter struct {
	cli *clientv3.Client
	limiter *rate.Limiter

}

func NewGreeter() (*greeter, error) {
	// 初始化etcd客户端
	cli, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 3,
	})
	imp := &greeter{
		cli: cli,
		limiter:rate.NewLimiter(100,10),
	}

	return imp, nil
}

// 注册服务到etcd
func (g *greeter) ServiceRegister(scheme, serviceName, addr string) {
	leaseGrantResp, err := g.cli.Grant(context.TODO(), 9)
	if err != nil {
		log.Fatal(err)
	}
	key := fmt.Sprintf("/%s/%s/%s", scheme, ServiceName, addr)
	// 执行put方法
	timeoutCtx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err = g.cli.Put(timeoutCtx, key, addr, clientv3.WithPrevKV(), clientv3.WithLease(leaseGrantResp.ID))
	if err != nil {
		// 注册失败直接panic
		panic(err)
	}
	go func() {
		for {
			ch, err := g.cli.KeepAlive(context.TODO(), leaseGrantResp.ID)
			if err != nil {
				fmt.Println(err)
			}
			select {
			case ka := <-ch:
				// 续租前过期，重新生成新的租约，中心续租
				if ka == nil {
					leaseGrantResp, err = g.cli.Grant(context.TODO(), 2)
					if err != nil {
						log.Fatal(err)
					}
					timeoutCtx, _ := context.WithTimeout(context.Background(), 1*time.Second)
					_, err = g.cli.Put(timeoutCtx, key, addr, clientv3.WithPrevKV(), clientv3.WithLease(leaseGrantResp.ID))
					if err != nil {
						// 注册失败直接panic
						panic(err)
					}
				}
				fmt.Println("ttl:", ka)
			}
		}
	}()
}

// 从etcd注销服务
func (g *greeter) ServiceLogout(scheme, serviceName, addr string) {
	key := fmt.Sprintf("/%s/%s/%s", scheme, ServiceName, addr)
	// 执行put方法
	timeoutCtx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := g.cli.Delete(timeoutCtx, key)
	if err != nil {
		// 注册失败直接panic
		panic(err)
	}
}
