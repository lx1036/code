ENDPOINTS := "http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379"
ENDPOINTS_AUTH := "https://127.0.0.1:42379,https://127.0.0.1:52379,https://127.0.0.1:62379"
CLUSTER := "etcd12379=http://127.0.0.1:12380,etcd22379=http://127.0.0.1:22380,etcd32379=http://127.0.0.1:32380"
ENDPOINT1 := "http://127.0.0.1:12379"
CLUSTER_TOKEN := "etcd-cluster-1"
PWD := $(shell pwd)

install:
	wget https://github.com/etcd-io/etcd/releases/download/v3.4.4/etcd-v3.4.4-darwin-amd64.zip && \
	tar etcd-v3.4.4-darwin-amd64.zip && cd etcd-v3.4.4-darwin-amd64 && \
	rm /usr/local/bin/etcd && rm /usr/local/bin/etcdctl && \
	mv ./etcd /usr/local/bin/ && mv ./etcdctl /usr/local/bin/

list:
	ETCDCTL_API=3 etcdctl member list \
		--write-out=table \
		--endpoints=https://127.0.0.1:42379 \
		--cacert $(PWD)/tls/ca.pem \
		--cert $(PWD)/tls/etcd-client.pem \
		--key $(PWD)/tls/etcd-client-key.pem

etcd-auth:
	etcd \
      --name=infra1 \
      --client-cert-auth=true \
      --cert-file=$(PWD)/tls/etcd.pem \
      --key-file=$(PWD)/tls/etcd-key.pem \
      --peer-client-cert-auth=true \
      --peer-cert-file=$(PWD)/tls/etcd.pem \
      --peer-key-file=$(PWD)/tls/etcd-key.pem \
      --trusted-ca-file=$(PWD)/tls/ca.pem \
      --peer-trusted-ca-file=$(PWD)/tls/ca.pem \
      --initial-advertise-peer-urls=https://127.0.0.1:42380 \
      --listen-peer-urls=https://127.0.0.1:42380 \
      --listen-client-urls=https://127.0.0.1:42379 \
      --advertise-client-urls=https://127.0.0.1:42379 \
      --initial-cluster-token=etcd-cluster-0 \
      --initial-cluster 'infra1=https://127.0.0.1:42380' \
      --initial-cluster-state=new \
      --data-dir=$(PWD)/infra2.etcd \
      --logger=zap \
      --log-outputs=stderr

etcd1-auth:
	etcd \
      --name=infra1 \
      --client-cert-auth=true \
      --cert-file=$(PWD)/tls/etcd.pem \
      --key-file=$(PWD)/tls/etcd-key.pem \
      --peer-client-cert-auth=true \
      --peer-cert-file=$(PWD)/tls/etcd.pem \
      --peer-key-file=$(PWD)/tls/etcd-key.pem \
      --trusted-ca-file=$(PWD)/tls/ca.pem \
      --peer-trusted-ca-file=$(PWD)/tls/ca.pem \
      --initial-advertise-peer-urls=https://127.0.0.1:42380 \
      --listen-peer-urls=https://127.0.0.1:42380 \
      --listen-client-urls=https://127.0.0.1:42379 \
      --advertise-client-urls=https://127.0.0.1:42379 \
      --initial-cluster-token=etcd-cluster-0 \
      --initial-cluster 'infra1=https://127.0.0.1:42380,infra2=https://127.0.0.1:52380,infra3=https://127.0.0.1:62380' \
      --initial-cluster-state=new \
      --data-dir=$(PWD)/infra1.etcd \
      --logger=zap \
      --log-outputs=stderr

etcd2-auth:
	etcd \
      --name=infra2 \
      --client-cert-auth=true \
      --cert-file=$(PWD)/tls/etcd.pem \
      --key-file=$(PWD)/tls/etcd-key.pem \
      --peer-client-cert-auth=true \
      --peer-cert-file=$(PWD)/tls/etcd.pem \
      --peer-key-file=$(PWD)/tls/etcd-key.pem \
      --trusted-ca-file=$(PWD)/tls/ca.pem \
      --peer-trusted-ca-file=$(PWD)/tls/ca.pem \
      --initial-advertise-peer-urls=https://127.0.0.1:52380 \
      --listen-peer-urls=https://127.0.0.1:52380 \
      --listen-client-urls=https://127.0.0.1:52379 \
      --advertise-client-urls=https://127.0.0.1:52379 \
      --initial-cluster-token=etcd-cluster-0 \
      --initial-cluster 'infra1=https://127.0.0.1:42380,infra2=https://127.0.0.1:52380,infra3=https://127.0.0.1:62380' \
      --initial-cluster-state=new \
      --data-dir=$(PWD)/infra2.etcd \
      --logger=zap \
      --log-outputs=stderr

etcd3-auth:
	etcd \
      --name=infra3 \
      --client-cert-auth=true \
      --cert-file=$(PWD)/tls/etcd.pem \
      --key-file=$(PWD)/tls/etcd-key.pem \
      --peer-client-cert-auth=true \
      --peer-cert-file=$(PWD)/tls/etcd.pem \
      --peer-key-file=$(PWD)/tls/etcd-key.pem \
      --trusted-ca-file=$(PWD)/tls/ca.pem \
      --peer-trusted-ca-file=$(PWD)/tls/ca.pem \
      --initial-advertise-peer-urls=https://127.0.0.1:62380 \
      --listen-peer-urls=https://127.0.0.1:62380 \
      --listen-client-urls=https://127.0.0.1:62379 \
      --advertise-client-urls=https://127.0.0.1:62379 \
      --initial-cluster-token=etcd-cluster-0 \
      --initial-cluster 'infra1=https://127.0.0.1:42380,infra2=https://127.0.0.1:52380,infra3=https://127.0.0.1:62380' \
      --initial-cluster-state=new \
      --data-dir=$(PWD)/infra3.etcd \
      --logger=zap \
      --log-outputs=stderr


etcd1:
	etcd --name etcd12379 \
      --listen-client-urls http://127.0.0.1:12379 \
      --advertise-client-urls http://127.0.0.1:12379 \
      --listen-peer-urls http://127.0.0.1:12380 \
      --initial-advertise-peer-urls http://127.0.0.1:12380 \
      --initial-cluster-token $(CLUSTER_TOKEN) \
      --initial-cluster $(CLUSTER) \
      --initial-cluster-state new \
      --enable-pprof \
      --logger=zap \
      --log-outputs=stderr

etcd2:
	etcd --name etcd22379 \
      --listen-client-urls http://127.0.0.1:22379 \
      --advertise-client-urls http://127.0.0.1:22379 \
      --listen-peer-urls http://127.0.0.1:22380 \
      --initial-advertise-peer-urls http://127.0.0.1:22380 \
      --initial-cluster-token $(CLUSTER_TOKEN) \
      --initial-cluster $(CLUSTER) \
      --initial-cluster-state new \
      --enable-pprof \
      --logger=zap \
      --log-outputs=stderr

etcd3:
	etcd --name etcd32379 \
      --listen-client-urls http://127.0.0.1:32379 \
      --listen-peer-urls http://127.0.0.1:32380 \
      --advertise-client-urls http://127.0.0.1:32379 \
      --initial-advertise-peer-urls http://127.0.0.1:32380 \
      --initial-cluster-token $(CLUSTER_TOKEN) \
      --initial-cluster $(CLUSTER) \
      --initial-cluster-state new \
      --enable-pprof \
      --logger=zap \
      --log-outputs=stderr

watch:
	etcdctl --endpoints=$(ENDPOINTS) watch foo

status:
	etcdctl --write-out=table --endpoints=$(ENDPOINTS) endpoint status

health:
	etcdctl --write-out=table --endpoints=$(ENDPOINTS) endpoint health

health-auth:
	etcdctl -w=table --endpoints=$(ENDPOINTS_AUTH) endpoint status  \
	--cacert $(PWD)/tls/ca.pem \
    --cert $(PWD)/tls/etcd-client.pem \
    --key $(PWD)/tls/etcd-client-key.pem

snapshot:
	etcdctl --endpoints=$(ENDPOINT1) snapshot save my.db
	etcdctl --write-out=table --endpoints=$(ENDPOINT1) snapshot status my.db


# etcd --name=etcd1 --data-dir=cluster1/etcd1 --initial-advertise-peer-urls=http://127.0.0.1:12380 --advertise-client-urls=http://127.0.0.1:12379 --listen-peer-urls=http://0.0.0.0:12380 --listen-client-urls=http://0.0.0.0:12379 --initial-cluster="etcd1=http://127.0.0.1:12380" --initial-cluster-token=abc123 --initial-cluster-state=new
# etcdctl member add etcd2 --peer-urls="http://127.0.0.1:22380"
# etcd --name=etcd2 --data-dir=cluster1/etcd2 --initial-advertise-peer-urls=http://127.0.0.1:22380 --advertise-client-urls=http://127.0.0.1:22379 --listen-peer-urls=http://0.0.0.0:22380 --listen-client-urls=http://0.0.0.0:22379 --initial-cluster="etcd1=http://127.0.0.1:12380,etcd2=http://127.0.0.1:22380" --initial-cluster-state=existing
# etcdctl member add etcd3 --peer-urls="http://127.0.0.1:32380"
# etcd --name=etcd3 --data-dir=cluster1/etcd3 --initial-advertise-peer-urls=http://127.0.0.1:32380 --advertise-client-urls=http://127.0.0.1:32379 --listen-peer-urls=http://0.0.0.0:32380 --listen-client-urls=http://0.0.0.0:32379 --initial-cluster="etcd1=http://127.0.0.1:12380,etcd2=http://127.0.0.1:22380,etcd3=http://127.0.0.1:32380" --initial-cluster-state=existing
# etcdctl member add etcd4 --peer-urls="http://127.0.0.1:42380" --learner
# etcd --name=etcd4 --data-dir=cluster1/etcd4 --initial-advertise-peer-urls=http://127.0.0.1:42380 --advertise-client-urls=http://127.0.0.1:42379 --listen-peer-urls=http://0.0.0.0:42380 --listen-client-urls=http://0.0.0.0:42379 --initial-cluster="etcd1=http://127.0.0.1:12380,etcd2=http://127.0.0.1:22380,etcd3=http://127.0.0.1:32380,etcd4=http://127.0.0.1:42380" --initial-cluster-state=existing

