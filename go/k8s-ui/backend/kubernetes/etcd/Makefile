ENDPOINTS := "http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379"
CLUSTER := "etcd12379=http://127.0.0.1:12380,etcd22379=http://127.0.0.1:22380,etcd32379=http://127.0.0.1:32380"
ENDPOINT1 := "http://127.0.0.1:12379"
CLUSTER_TOKEN := "etcd-cluster-1"

install:
	wget https://github.com/etcd-io/etcd/releases/download/v3.4.4/etcd-v3.4.4-darwin-amd64.zip && \
	tar etcd-v3.4.4-darwin-amd64.zip && cd etcd-v3.4.4-darwin-amd64 && \
	rm /usr/local/bin/etcd && rm /usr/local/bin/etcdctl && \
	mv ./etcd /usr/local/bin/ && mv ./etcdctl /usr/local/bin/

tls:
	etcd --cert-file ./kubernetes.pem --key-file ./kubernetes-key.pem --trusted-ca-file ./ca.pem

etcd1:
	etcd --name etcd12379 \
      --listen-client-urls http://127.0.0.1:12379 \
      --advertise-client-urls http://127.0.0.1:12379 \
      --listen-peer-urls http://127.0.0.1:12380 \
      --initial-advertise-peer-urls http://127.0.0.1:12380 \
      --initial-cluster-token $(CLUSTER_TOKEN) \
      --initial-cluster $(CLUSTER) \
      --initial-cluster-state new \
      --enable-pprof --logger=zap --log-outputs=stderr

etcd2:
	etcd --name etcd22379 \
      --listen-client-urls http://127.0.0.1:22379 \
      --advertise-client-urls http://127.0.0.1:22379 \
      --listen-peer-urls http://127.0.0.1:22380 \
      --initial-advertise-peer-urls http://127.0.0.1:22380 \
      --initial-cluster-token $(CLUSTER_TOKEN) \
      --initial-cluster $(CLUSTER) \
      --initial-cluster-state new \
      --enable-pprof --logger=zap --log-outputs=stderr

etcd3:
	etcd --name etcd32379 \
      --listen-client-urls http://127.0.0.1:32379 \
      --listen-peer-urls http://127.0.0.1:32380 \
      --advertise-client-urls http://127.0.0.1:32379 \
      --initial-advertise-peer-urls http://127.0.0.1:32380 \
      --initial-cluster-token $(CLUSTER_TOKEN) \
      --initial-cluster $(CLUSTER) \
      --initial-cluster-state new \
      --enable-pprof --logger=zap --log-outputs=stderr

watch:
	etcdctl --endpoints=$(ENDPOINTS) watch foo

status:
	etcdctl --write-out=table --endpoints=$(ENDPOINTS) endpoint status

health:
	etcdctl --write-out=table --endpoints=$(ENDPOINTS) endpoint health

snapshot:
	etcdctl --endpoints=$(ENDPOINT1) snapshot save my.db
	etcdctl --write-out=table --endpoints=$(ENDPOINT1) snapshot status my.db
