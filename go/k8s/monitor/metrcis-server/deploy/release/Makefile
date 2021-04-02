


# @see https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.4.2/components.yaml
# 可以参考官网：[安装一个扩展的 API 服务器来使用聚合层](https://kubernetes.io/zh/docs/tasks/extend-kubernetes/setup-extension-api-server/)
install:
	kubectl apply --kustomize .

cleanup:
	kubectl delete --kustomize .

check:
	kubectl get --raw /apis/metrics.k8s.io/v1beta1
