package common

import (
	"crypto/tls"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strings"
	"sync"
	"time"
)

var (
	EtcdClient = &sync.Map{}
)

func NewEtcdClient(config *EtcdServer) (*clientv3.Client, error) {
	var (
		client    *clientv3.Client
		tlsConfig *tls.Config
		err       error
	)
	if config.TLSEnable {
		tlsInfo := transport.TLSInfo{
			CertFile:      config.CertFile,
			KeyFile:       config.KeyFile,
			TrustedCAFile: config.CAFile,
		}
		tlsConfig, err = tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}
	}

	client, err = clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(config.Endpoints, ","),
		DialTimeout: time.Second * 5,
		TLS:         tlsConfig,
	})
	if err != nil {
		return nil, err
	}

	EtcdClient.Store(config.Name, client)

	return client, nil
}
