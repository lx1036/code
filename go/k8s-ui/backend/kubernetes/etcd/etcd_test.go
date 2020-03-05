package etcd

import (
	"context"
	"flag"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"go.etcd.io/etcd/pkg/transport"
	"google.golang.org/grpc/grpclog"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

/**
go run etcd.go --endpoint localhost:12379
*/


func TestClientv3(test *testing.T) {
	endpoint := flag.String("endpoint", "localhost:2379", "talk with client")
	flag.Parse()

	clientv3.SetLogger(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))

	client, err := clientv3.New(clientv3.Config{
		Endpoints:            []string{*endpoint},
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	response, err := client.Put(ctx, "foo1", "bar1")
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

// etcd --cert-file ./kubernetes.pem --key-file ./kubernetes-key.pem --trusted-ca-file ./ca.pem
func TestClientv3WithTLS(test *testing.T) {
	abs, _ := filepath.Abs(".")
	tlsInfo := transport.TLSInfo{
		CertFile:            abs + "/kubernetes.pem",
		KeyFile:             abs + "/kubernetes-key.pem",
		TrustedCAFile:       abs + "/ca.pem",
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		panic(err)
	}

	fmt.Println(tlsConfig.MaxVersion)

	endpoint := flag.String("endpoint", "localhost:2379", "talk with client")
	flag.Parse()

	clientv3.SetLogger(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))

	client, err := clientv3.New(clientv3.Config{
		Endpoints:            []string{*endpoint},
		DialTimeout:          time.Second * 5,
		TLS:                  tlsConfig,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	response, err := client.Get(ctx, "foo1")
	if err != nil {
		panic(err)
	}
	fmt.Println(*response)

	watch := client.Watch(ctx, "foo1")
	for w := range watch {
		for _, event := range w.Events {
			fmt.Printf("%s %q:%q\n", event.Type, event.Kv.Key, event.Kv.Value)
		}
	}
}
