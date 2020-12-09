package balancer

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
	"strings"
	"sync"
	"time"
)

const (
	ExampleScheme = "test"
)

var cli *clientv3.Client
var once sync.Once

type ETCDResolverBuilder struct {
	ETCDHost string
}

func (e *ETCDResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	once.Do(func() {
		clientConfig := clientv3.Config{
			Endpoints: []string{"127.0.0.1:2379"},
		}
		cli, _ = clientv3.New(clientConfig)
		// 连接检查
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_, err := cli.Status(timeoutCtx, clientConfig.Endpoints[0])
		if err != nil {
			panic(err)
		}
	})
	r := &ETCDResolver{
		target: target,
		cc:     cc,
	}
	key := fmt.Sprintf("/%s/%s/", target.Scheme, target.Endpoint)
	// 开启解析
	r.start(key)
	return r, nil
}

func (e *ETCDResolverBuilder) Scheme() string {
	return ExampleScheme
}

type ETCDResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrStores map[string][]string
}

func (e *ETCDResolver) ResolveNow(o resolver.ResolveNowOptions) {
	fmt.Println("ResolveNow")
}
func (e *ETCDResolver) Close() {
	fmt.Println("close")
}

func (e *ETCDResolver) start(serviceName string) {
	var addrList []resolver.Address
	// 前缀查询
	if result, err := cli.Get(context.Background(), serviceName, clientv3.WithPrefix()); err == nil {
		for i := range result.Kvs {
			fmt.Println("key：", string(result.Kvs[i].Key))
			fmt.Println("value：", string(result.Kvs[i].Value))
			fmt.Println("strings = ", strings.TrimPrefix(string(result.Kvs[i].Key), serviceName))
			addrList = append(addrList, resolver.Address{Addr: string(result.Kvs[i].Value)})
		}
	}
	e.cc.UpdateState(resolver.State{Addresses: addrList})
	go e.watch(serviceName, addrList)
}

func (e *ETCDResolver) watch(serviceName string, addrList []resolver.Address) {
	addrMapper := make(map[string]int)
	for _, addr := range addrList {
		addrMapper[addr.Addr] = 1
	}
	// 前缀查询
	ch := cli.Watch(context.TODO(), serviceName, clientv3.WithPrefix())
	for {
		select {
		case c := <-ch:
			for _, event := range c.Events {
				addr := strings.TrimPrefix(string(event.Kv.Key), serviceName)
				switch event.Type {
				case mvccpb.PUT:
					fmt.Println("put addr = ",addr)
					if _, ok := addrMapper[addr]; !ok {
						addrList = append(addrList, resolver.Address{Addr: addr})
						e.cc.UpdateState(resolver.State{Addresses: addrList})
						addrMapper[addr] = 1
					}
				case mvccpb.DELETE:
					if _, ok := addrMapper[addr]; ok {
						for i := range addrList {
							if addrList[i].Addr == addr {
								addrList[i] = addrList[len(addrList)-1]
								addrList = addrList[:len(addrList)-1]
								delete(addrMapper, addr)
								fmt.Println("delete addr = ",addr)
								break
							}
						}
						e.cc.UpdateState(resolver.State{Addresses: addrList})
					}
				}
			}
		}
	}
}
