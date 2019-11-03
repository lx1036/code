package main

import (
    "context"
    "fmt"
    "github.com/coreos/etcd/clientv3"
    "github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
    "log"
    "time"
)

func main() {
    client, err := clientv3.New(clientv3.Config{
        Endpoints:            []string{"localhost:2379"},
        AutoSyncInterval:     0,
        DialTimeout:          0,
        DialKeepAliveTime:    0,
        DialKeepAliveTimeout: 0,
        MaxCallSendMsgSize:   0,
        MaxCallRecvMsgSize:   0,
        TLS:                  nil,
        Username:             "",
        Password:             "",
        RejectOldCluster:     false,
        DialOptions:          nil,
        LogConfig:            nil,
        Context:              nil,
        PermitWithoutStream:  false,
    })
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Second * 3)
    response , err := client.Put(ctx, "foo1", "bar1")
    defer cancel()
    if err != nil {
        switch err {
        case context.Canceled:
            log.Fatalf("ctx is canceled by another routine: %v", err)
        case context.DeadlineExceeded:
            log.Fatalf("ctx is attached with a deadline is exceeded: %v", err)
        case rpctypes.ErrEmptyKey:
            log.Fatalf("client-side error: %v", err)
        default:
            log.Fatalf("bad cluster endpoints, which are not etcd servers: %v", err)
        }
    }
    
    fmt.Println(*response)
    
    getResponse, err := client.Get(ctx, "foo")
    if err != nil {
        switch err {
        case context.Canceled:
            log.Fatalf("ctx is canceled by another routine: %v", err)
        case context.DeadlineExceeded:
            log.Fatalf("ctx is attached with a deadline is exceeded: %v", err)
        case rpctypes.ErrEmptyKey:
            log.Fatalf("client-side error: %v", err)
        default:
            log.Fatalf("bad cluster endpoints, which are not etcd servers: %v", err)
        }
    }
    
    fmt.Println(*getResponse)
}





