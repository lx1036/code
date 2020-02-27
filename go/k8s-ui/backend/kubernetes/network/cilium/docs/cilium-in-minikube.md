
# Cilium 是什么？
Cilium是k8s的网络插件的一种可选项，可给pod安全分配网络连接。

# Install Cilium in Minikube
```shell script
minikube start --network-plugin=cni --memory=4096 # Create a minikube cluster
minikube ssh -- sudo mount bpffs -t bpf /sys/fs/bpf # Mount the BPF filesystem
kubectl create -f https://raw.githubusercontent.com/cilium/cilium/v1.7/install/kubernetes/quick-install.yaml # Install Cilium as DaemonSet into your new Kubernetes cluster. 
kubectl -n kube-system get pods --watch # watch
kubectl apply -f https://raw.githubusercontent.com/cilium/cilium/v1.7/examples/kubernetes/connectivity-check/connectivity-check.yaml # validate
kubectl apply -f https://raw.githubusercontent.com/cilium/hubble/master/tutorials/deploy-hubble-servicemap/hubble-all-minikube.yaml # install Hubble
```



