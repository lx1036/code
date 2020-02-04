

etcd --name etcd12379 \
  --listen-client-urls http://127.0.0.1:12379 \
  --advertise-client-urls http://127.0.0.1:12379 \
  --listen-peer-urls http://127.0.0.1:12380 \
  --initial-advertise-peer-urls http://127.0.0.1:12380 \
  --initial-cluster-token etcd-cluster-1 \
  --initial-cluster 'etcd12379=http://127.0.0.1:12380,etcd22379=http://127.0.0.1:22380,etcd32379=http://127.0.0.1:32380' \
  --initial-cluster-state new \
  --enable-pprof --logger=zap --log-outputs=stderr
  
etcd --name etcd22379 \
  --listen-client-urls http://127.0.0.1:22379 \
  --advertise-client-urls http://127.0.0.1:22379 \
  --listen-peer-urls http://127.0.0.1:22380 \
  --initial-advertise-peer-urls http://127.0.0.1:22380 \
  --initial-cluster-token etcd-cluster-1 \
  --initial-cluster 'etcd12379=http://127.0.0.1:12380,etcd22379=http://127.0.0.1:22380,etcd32379=http://127.0.0.1:32380' \
  --initial-cluster-state new \
  --enable-pprof --logger=zap --log-outputs=stderr

etcd --name etcd32379 \
  --listen-client-urls http://127.0.0.1:32379 \
  --advertise-client-urls http://127.0.0.1:32379 \
  --listen-peer-urls http://127.0.0.1:32380 \
  --initial-advertise-peer-urls http://127.0.0.1:32380 \
  --initial-cluster-token etcd-cluster-1 \
  --initial-cluster 'etcd12379=http://127.0.0.1:12380,etcd22379=http://127.0.0.1:22380,etcd32379=http://127.0.0.1:32380' \
  --initial-cluster-state new \
  --enable-pprof --logger=zap --log-outputs=stderr
