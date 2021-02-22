package lease

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/clientv3"
	"k8s.io/klog/v2"
)

func ExampleLease_keepAliveOnce() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 10,
	})
	if err != nil {
		klog.Fatal(err)
	}
	defer cli.Close()

	resp, err := cli.Grant(context.TODO(), 50)
	if err != nil {
		klog.Fatal(err)
	}

	_, err = cli.Put(context.TODO(), "foo", "bar", clientv3.WithLease(resp.ID))
	if err != nil {
		klog.Fatal(err)
	}

	// to renew the lease only once
	ka, kaerr := cli.KeepAliveOnce(context.TODO(), resp.ID)
	if kaerr != nil {
		klog.Fatal(kaerr)
	}

	fmt.Println("ttl:", ka.TTL)
	// Output: ttl: 50
}

func ExampleLease_keepAlive() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 10,
	})
	if err != nil {
		klog.Fatal(err)
	}
	defer cli.Close()

	resp, err := cli.Grant(context.TODO(), 5)
	if err != nil {
		klog.Fatal(err)
	}

	_, err = cli.Put(context.TODO(), "foo", "bar", clientv3.WithLease(resp.ID))
	if err != nil {
		klog.Fatal(err)
	}

	// the key 'foo' will be kept forever
	ch, kaerr := cli.KeepAlive(context.TODO(), resp.ID)
	if kaerr != nil {
		klog.Fatal(kaerr)
	}

	ka := <-ch
	fmt.Println("ttl:", ka.TTL)
	// Output: ttl: 5
}
