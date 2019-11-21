for (( i = 0; i < 3; i++ )); do
    gcloud compute instances create worker-${i} --async --boot-disk-size 200GB \
      --can-ip-forward --image-family ubuntu-1804-lts \
      --image-project ubuntu-os-cloud --machine-type n1-standard-1 \
      --private-network-ip 10.240.0.2${i} --metadata pod-cidr=10.200.${i}.0/24 \
      --scopes compute-rw,storage-ro,service-management,service-control,logging-write,monitoring \
      --subnet kubernetes --tags kubernetes-the-hard-way,controller
done
