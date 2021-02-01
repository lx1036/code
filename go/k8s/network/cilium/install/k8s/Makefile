





install:
	minikube ssh -- sudo mount bpffs -t bpf /sys/fs/bpf
	kubectl apply --kustomize .

cleanup:
	kubectl delete --kustomize .



