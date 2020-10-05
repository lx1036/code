
up:
	rke up --config rancher-k8s.yaml
	kubectl --kubeconfig ./kube_config_rancher-k8s.yaml get nodes
	kubectl --kubeconfig ./kube_config_rancher-k8s.yaml apply -f ./cilium.yaml
	kubectl --kubeconfig ./kube_config_rancher-k8s.yaml get pods -A --watch

nginx:
	kubectl --kubeconfig ./kube_config_rancher-k8s.yaml apply -f ./nginx.yaml
