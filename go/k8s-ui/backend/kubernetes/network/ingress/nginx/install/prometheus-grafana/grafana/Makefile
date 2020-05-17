
PORT := $(shell kubectl get svc grafana-server -o=jsonpath='{.spec.ports[0].nodePort}' -n ingress-nginx)
IP := $(shell minikube ip)

chrome:
	open http://$(IP):$(PORT)

install:
	kubectl apply --kustomize .
