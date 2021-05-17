
# generic server
代码在：staging/src/k8s.io/apiserver



apiserver 启动参数：
```json

"Entrypoint": [
    "/opt/rke-tools/entrypoint.sh",
    "kube-apiserver",
    "--etcd-keyfile=/etc/kubernetes/ssl/kube-node-key.pem",
    "--kubelet-client-key=/etc/kubernetes/ssl/kube-apiserver-key.pem",
    "--authentication-token-webhook-config-file=/etc/kubernetes/kube-api-authn-webhook.yaml",
    "--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
    "--service-account-lookup=true",
    "--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
    "--audit-log-maxage=30",
    "--client-ca-file=/etc/kubernetes/ssl/kube-ca.pem",
    "--requestheader-client-ca-file=/etc/kubernetes/ssl/kube-apiserver-requestheader-ca.pem",
    "--service-account-key-file=/etc/kubernetes/ssl/kube-service-account-token-key.pem",
    "--service-cluster-ip-range=192.168.0.0/16",
    "--authentication-token-webhook-cache-ttl=5s",
    "--insecure-port=0",
    "--bind-address=0.0.0.0",
    "--proxy-client-cert-file=/etc/kubernetes/ssl/kube-apiserver-proxy-client.pem",
    "--enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,NodeRestriction,Priority,TaintNodesByCondition,PersistentVolumeClaimResize",
    "--anonymous-auth=false",
    "--audit-log-maxbackup=10",
    "--audit-log-format=json",
    "--audit-policy-file=/etc/kubernetes/audit-policy.yaml",
    "--etcd-cafile=/etc/kubernetes/ssl/kube-ca.pem",
    "--etcd-certfile=/etc/kubernetes/ssl/kube-node.pem",
    "--etcd-prefix=/registry",
    "--tls-cert-file=/etc/kubernetes/ssl/kube-apiserver.pem",
    "--tls-private-key-file=/etc/kubernetes/ssl/kube-apiserver-key.pem",
    "--requestheader-username-headers=X-Remote-User",
    "--secure-port=6443",
    "--service-node-port-range=30000-32767",
    "--requestheader-extra-headers-prefix=X-Remote-Extra-",
    "--audit-log-maxsize=100",
    "--cloud-provider=",
    "--etcd-servers=https://10.1.2.3:2379,https://10.1.2.4:2379,https://10.1.2.5:2379",
    "--proxy-client-key-file=/etc/kubernetes/ssl/kube-apiserver-proxy-client-key.pem",
    "--runtime-config=authorization.k8s.io/v1beta1=true",
    "--authorization-mode=Node,RBAC",
    "--audit-log-path=/var/log/kube-audit/audit-log.json",
    "--enable-aggregator-routing=true",
    "--requestheader-allowed-names=kube-apiserver-proxy-client",
    "--allow-privileged=true",
    "--storage-backend=etcd3",
    "--profiling=false",
    "--kubelet-client-certificate=/etc/kubernetes/ssl/kube-apiserver.pem",
    "--requestheader-group-headers=X-Remote-Group",
    "--advertise-address=10.1.2.3"
]

```



## 参考文献
**[Kubernetes API Server Generic API Server 架构设计源码阅读](https://cloudnative.to/blog/kubernetes-apiserver-generic-api-server/)**


