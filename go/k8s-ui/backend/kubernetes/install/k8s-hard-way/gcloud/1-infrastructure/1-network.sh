
# network
gcloud compute networks create kubernetes-the-hard-way --subnet-mode custom #vpc

gcloud compute networks subnets create kubernetes \
  --network kubernetes-the-hard-way \
  --range 10.240.0.0/24 # subnet 子网

gcloud compute firewall-rules create kubernetes-the-hard-way-allow-internal \
  --allow tcp,udp,icmp \
  --network kubernetes-the-hard-way \
  --source-ranges 10.240.0.0/24,10.200.0.0/16 # firewall-rules 内网防火墙
gcloud compute firewall-rules create kubernetes-the-hard-way-allow-external \
  --allow tcp:22,tcp:6443,icmp \
  --network kubernetes-the-hard-way \
  --source-ranges 0.0.0.0/0 # firewall-rules 外网防火墙，容许ssh,https,icmp访问
gcloud compute firewall-rules list --filter="network:kubernetes-the-hard-way" # 列出防火墙规则

gcloud compute addresses create kubernetes-the-hard-way \
  --region $(gcloud config get-value compute/region) # External IP addresses 分配静态IP地址
gcloud compute addresses list --filter="name=('kubernetes-the-hard-way')"

