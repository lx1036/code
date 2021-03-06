
# network
gcloud compute networks create kubernetes-the-hard-way --subnet-mode custom #vpc
gcloud compute networks subnets create kubernetes --network kubernetes-the-hard-way --range 10.240.0.0/24 # subnet
gcloud compute firewall-rules create kubernetes-the-hard-way-allow-internal --allow tcp,udp,icmp \
  --network kubernetes-the-hard-way \
  --source-ranges 10.240.0.0/24,10.200.0.0/16 # firewall-rules
gcloud compute firewall-rules create kubernetes-the-hard-way-allow-external --allow tcp:22,tcp:6443,icmp \
  --network kubernetes-the-hard-way \
  --source-ranges 0.0.0.0/0 # firewall-rules
gcloud compute addresses create kubernetes-the-hard-way --region $(gcloud config get-value compute/region) # address

# VM
## gcp master-instances
for (( i = 0; i < 3; i++ )); do
    gcloud compute instances create controller-${i} --async --boot-disk-size 200GB \
      --can-ip-forward --image-family ubuntu-1804-lts \
      --image-project ubuntu-os-cloud --machine-type n1-standard-1 \
      --private-network-ip 10.240.0.1${i} \
      --scopes compute-rw,storage-ro,service-management,service-control,logging-write,monitoring \
      --subnet kubernetes --tags kubernetes-the-hard-way,controller
done
## gcp worker-instances
for (( i = 0; i < 3; i++ )); do
    gcloud compute instances create worker-${i} --async --boot-disk-size 200GB \
      --can-ip-forward --image-family ubuntu-1804-lts \
      --image-project ubuntu-os-cloud --machine-type n1-standard-1 \
      --private-network-ip 10.240.0.2${i} --metadata pod-cidr=10.200.${i}.0/24 \
      --scopes compute-rw,storage-ro,service-management,service-control,logging-write,monitoring \
      --subnet kubernetes --tags kubernetes-the-hard-way,worker
done
## check instances
gcloud compute instances list
## ssh vm
gcloud compute ssh master-0
gcloud compute ssh master-1
gcloud compute ssh master-2
gcloud compute ssh worker-0
gcloud compute ssh worker-1
gcloud compute ssh worker-2
