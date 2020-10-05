
# rke
```shell script
brew install rke
```

## 使用rke安装k8s
(1)准备两台centos机器，安装docker，并初始化机器
```shell script
wget -O /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-7.repo
sed -i -e '/mirrors.cloud.aliyuncs.com/d' -e '/mirrors.aliyuncs.com/d' /etc/yum.repos.d/CentOS-Base.repo
yum makecache
yum install -y yum-utils
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum install docker-ce docker-ce-cli containerd.io -y
systemctl enable docker
#注意docker 跟目录设置 "graph": "/data/docker", 改为实际的registry ip:port  "insecure-registries": ["10.19.214.141"],
cat <<EOF | sudo tee /etc/docker/daemon.json
{
  "registry-mirrors": ["https://wvtedxym.mirror.aliyuncs.com"],
  "oom-score-adjust": -1000,
  "graph": "/data/docker",
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
    },
  "live-restore": true,
  "init": true
}
EOF
systemctl daemon-reload
systemctl restart docker

# 关闭swap
# 临时禁用
swapoff -a
# 永久禁用
# vi /etc/fstab
sed -i 's/.*swap.*/#&/' /etc/fstab
# 设置docker用户组
useradd dockeruser
usermod -aG docker liuxiang3
# 关闭Selinux，设置SELINUX=disabled
vi /etc/sysconfig/selinux
# 设置IPV4转发
vi /etc/sysctl.conf
net.ipv4.ip_forward = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
# 设置IPV4转发生效
sudo sysctl -p
#分发密钥
ssh-copy-id liuxiang3@xxx.xxx.xxx.xxx
```
(2)rke客户机上运行命令，安装k8s
(2.1) 本地安装rke
```shell script
brew install rke
```

(2.2) 配置rke配置文件cluster.yaml，主要需要关注service_cluster_ip_range、service_node_port_range这几个和网络相关的字段。
具体字段含义可看文档 https://rancher.com/docs/rke/latest/en/example-yamls/。网络使用cilium，daemonset部署，这里network.plugin字段
设置为`none`：
```yaml
###
# https://rancher.com/docs/rke/latest/en/example-yamls/
###

# If you intened to deploy Kubernetes in an air-gapped environment,
# please consult the documentation on how to configure custom RKE images.
nodes:
  - address: p46282v.hulk.zzzc.qihoo.net
    port: "22"
    internal_address: 10.174.224.180
    role:
      - controlplane
      - worker
      - etcd
    hostname_override: ""
    user: liuxiang3
    docker_socket: /var/run/docker.sock
    ssh_key: ""
    ssh_key_path: ~/.ssh/id_rsa
    ssh_cert: ""
    ssh_cert_path: ""
    labels: {}
    taints: []
  - address: p46284v.hulk.shbt.qihoo.net
    port: "22"
    internal_address: 10.202.148.133
    role:
      - worker
    hostname_override: ""
    user: liuxiang3
    docker_socket: /var/run/docker.sock
    ssh_key: ""
    ssh_key_path: ~/.ssh/id_rsa
    ssh_cert: ""
    ssh_cert_path: ""
    labels: {}
    taints: []
services:
  etcd:
    image: ""
    extra_args: {}
    extra_binds: []
    extra_env: []
    win_extra_args: {}
    win_extra_binds: []
    win_extra_env: []
    external_urls: []
    ca_cert: ""
    cert: ""
    key: ""
    path: ""
    uid: 0
    gid: 0
    snapshot: null
    retention: ""
    creation: ""
    backup_config: null
  kube-api:
    image: ""
    extra_args: {}
    extra_binds: []
    extra_env: []
    win_extra_args: {}
    win_extra_binds: []
    win_extra_env: []
    # IP range for any services created on Kubernetes
    # This must match the service_cluster_ip_range in kube-controller
    service_cluster_ip_range: 192.168.0.0/16
    service_node_port_range: "30000-32767"
    pod_security_policy: false
    always_pull_images: false
    secrets_encryption_config: null
    audit_log: null
    admission_configuration: null
    event_rate_limit: null
  kube-controller:
    image: ""
    extra_args: {}
    extra_binds: []
    extra_env: []
    win_extra_args: {}
    win_extra_binds: []
    win_extra_env: []
    # CIDR pool used to assign IP addresses to pods in the cluster
    cluster_cidr: 172.20.0.0/16
    # IP range for any services created on Kubernetes
    # This must match the service_cluster_ip_range in kube-api
    service_cluster_ip_range: 192.168.0.0/16
  scheduler:
    image: ""
    extra_args: {}
    extra_binds: []
    extra_env: []
    win_extra_args: {}
    win_extra_binds: []
    win_extra_env: []
  kubelet:
    image: ""
    extra_args: {}
    extra_binds: []
    extra_env: []
    win_extra_args: {}
    win_extra_binds: []
    win_extra_env: []
    cluster_domain: cluster.local
    infra_container_image: ""
    cluster_dns_server: 192.168.0.2
    fail_swap_on: false
    generate_serving_certificate: false
  kubeproxy:
    image: ""
    extra_args: {}
    extra_binds: []
    extra_env: []
    win_extra_args: {}
    win_extra_binds: []
    win_extra_env: []
network:
  plugin: none
  options: {}
  mtu: 0
  node_selector: {}
  update_strategy: null
authentication:
  strategy: x509
  sans: []
  webhook: null
addons: ""
addons_include: []
system_images:
  etcd: rancher/coreos-etcd:v3.4.3-rancher1
  alpine: rancher/rke-tools:v0.1.64
  nginx_proxy: rancher/rke-tools:v0.1.64
  cert_downloader: rancher/rke-tools:v0.1.64
  kubernetes_services_sidecar: rancher/rke-tools:v0.1.64
  kubedns: rancher/k8s-dns-kube-dns:1.15.2
  dnsmasq: rancher/k8s-dns-dnsmasq-nanny:1.15.2
  kubedns_sidecar: rancher/k8s-dns-sidecar:1.15.2
  kubedns_autoscaler: rancher/cluster-proportional-autoscaler:1.7.1
  coredns: rancher/coredns-coredns:1.6.9
  coredns_autoscaler: rancher/cluster-proportional-autoscaler:1.7.1
  nodelocal: rancher/k8s-dns-node-cache:1.15.7
  kubernetes: rancher/hyperkube:v1.18.8-rancher1
  flannel: rancher/coreos-flannel:v0.12.0
  flannel_cni: rancher/flannel-cni:v0.3.0-rancher6
  calico_node: rancher/calico-node:v3.13.4
  calico_cni: rancher/calico-cni:v3.13.4
  calico_controllers: rancher/calico-kube-controllers:v3.13.4
  calico_ctl: rancher/calico-ctl:v3.13.4
  calico_flexvol: rancher/calico-pod2daemon-flexvol:v3.13.4
  canal_node: rancher/calico-node:v3.13.4
  canal_cni: rancher/calico-cni:v3.13.4
  canal_flannel: rancher/coreos-flannel:v0.12.0
  canal_flexvol: rancher/calico-pod2daemon-flexvol:v3.13.4
  weave_node: weaveworks/weave-kube:2.6.4
  weave_cni: weaveworks/weave-npc:2.6.4
  pod_infra_container: rancher/pause:3.1
  ingress: rancher/nginx-ingress-controller:nginx-0.32.0-rancher1
  ingress_backend: rancher/nginx-ingress-controller-defaultbackend:1.5-rancher1
  metrics_server: rancher/metrics-server:v0.3.6
  windows_pod_infra_container: rancher/kubelet-pause:v0.1.4
ssh_key_path: ~/.ssh/id_rsa
ssh_cert_path: ""
ssh_agent_auth: false
authorization:
  mode: rbac
  options: {}
ignore_docker_version: null
kubernetes_version: ""
private_registries: []
ingress:
  provider: ""
  options: {}
  node_selector: {}
  extra_args: {}
  dns_policy: ""
  extra_envs: []
  extra_volumes: []
  extra_volume_mounts: []
  update_strategy: null
cluster_name: ""
cloud_provider:
  name: ""
prefix_path: ""
win_prefix_path: ""
addon_job_timeout: 0
bastion_host:
  address: ""
  port: ""
  user: ""
  ssh_key: ""
  ssh_key_path: ""
  ssh_cert: ""
  ssh_cert_path: ""
monitoring:
  provider: ""
  options: {}
  node_selector: {}
  update_strategy: null
  replicas: null
restore:
  restore: false
  snapshot_name: ""
dns: null
```

(2.3) 运行 `rke up --config rancher-k8s.yaml`

(3)部署cilium
部署yaml主要来自于官方文档 https://github.com/cilium/cilium/blob/master/install/kubernetes/quick-install.yaml
```shell script
kubectl --kubeconfig ./kube_config_rancher-k8s.yaml apply -f ./cilium.yaml
```

(4)测试
```shell script
kubectl --kubeconfig ./kube_config_rancher-k8s.yaml apply -f ./nginx.yaml
```
