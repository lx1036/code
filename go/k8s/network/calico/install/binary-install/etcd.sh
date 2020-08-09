
CLUSTER_TOKEN := "etcd-cluster-1"
CLUSTER := "etcd2379=http://0.0.0.0:2380"

etcd --name etcd2379 \
      --listen-client-urls http://0.0.0.0:2379 \
      --advertise-client-urls http://0.0.0.0:2379 \
      --listen-peer-urls http://0.0.0.0:2380 \
      --initial-advertise-peer-urls http://0.0.0.0:2380 \
      --initial-cluster-token $(CLUSTER_TOKEN) \
      --initial-cluster $(CLUSTER) \
      --initial-cluster-state new \
      --enable-pprof \
      --logger=zap \
      --log-outputs=stderr
