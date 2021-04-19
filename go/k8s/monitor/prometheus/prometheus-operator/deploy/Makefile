





# https://prometheus-operator.dev/docs/prologue/quick-start/




IP := $(shell minikube ip)
PORT := $(shell kubectl get svc nginx-demo-1 -o=jsonpath='{.spec.ports[0].nodePort}')


minikube:
	minikube start --kubernetes-version=v1.20.0 --memory=6g --bootstrapper=kubeadm \
	--extra-config=kubelet.authentication-token-webhook=true --extra-config=kubelet.authorization-mode=Webhook \
	--extra-config=scheduler.address=0.0.0.0 --extra-config=controller-manager.address=0.0.0.0 \
	--image-repository=registry.cn-hangzhou.aliyuncs.com/google_containers --image-mirror-country='cn'

install:
	kubectl apply --kustomize .

