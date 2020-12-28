
IP := $(shell minikube ip)
PORT := $(shell kubectl get svc nginx-demo-1 -o=jsonpath='{.spec.ports[0].nodePort}')

install:
	kubectl apply --kustomize .

cleanup:
	kubectl delete --kustomize .

install-raw:
	kubectl apply -f ./crd.yaml
	kubectl apply -f ./configmap.yaml
	kubectl apply -f ./serviceaccount.yaml
	kubectl apply -f ./clusterrole.yaml
	kubectl apply -f ./deployment.yaml
	kubectl apply -f ./daemonset.yaml

cleanup-raw:
	kubectl delete -f ./daemonset.yaml
	kubectl delete -f ./deployment.yaml
	kubectl delete -f ./clusterrole.yaml
	kubectl delete -f ./serviceaccount.yaml
	kubectl delete -f ./configmap.yaml
	kubectl delete -f ./crd.yaml

chrome:
	open http://$(IP):$(PORT)

nodes:
	DATASTORE_TYPE=kubernetes KUBECONFIG=~/.kube/config calicoctl get nodes


minikube-etcd:
	kubectl exec etcd-minikube -n kube-system -- etcdctl etcdctl get / \
		--cert=/var/lib/minikube/certs/etcd/server.crt \
		--key=/var/lib/minikube/certs/etcd/server.key \
		--cacert=/var/lib/minikube/certs/etcd/ca.crt \
		--prefix  --keys-only | grep calico

node:
	kubectl exec etcd-minikube -n kube-system -- etcdctl --cert=/var/lib/minikube/certs/etcd/server.crt \
		--key=/var/lib/minikube/certs/etcd/server.key --cacert=/var/lib/minikube/certs/etcd/ca.crt \
		get /registry/crd.projectcalico.org/blockaffinities/minikube-10-217-120-64-26

ipam:
	kubectl exec etcd-minikube -n kube-system -- etcdctl --cert=/var/lib/minikube/certs/etcd/server.crt \
		--key=/var/lib/minikube/certs/etcd/server.key --cacert=/var/lib/minikube/certs/etcd/ca.crt \
		get /registry/crd.projectcalico.org/ipamblocks/10-217-120-64-26

