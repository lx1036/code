
IP := $(shell minikube ip)
PORT := $(shell kubectl get svc prometheus-server -o=jsonpath='{.spec.ports[0].nodePort}' -n ingress-nginx)

chrome:
	open http://$(IP):$(PORT)

install:
	kubectl apply --kustomize .
