


# @see https://github.com/kubernetes-sigs/prometheus-adapter/blob/master/deploy/README.md
# 可以参考官网：[安装一个扩展的 API 服务器来使用聚合层](https://kubernetes.io/zh/docs/tasks/extend-kubernetes/setup-extension-api-server/)
# https://www.qikqiak.com/post/k8s-hpa-usage/

install:
	kubectl apply --kustomize .

cleanup:
	kubectl delete --kustomize .

check:
	kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1 # http://127.0.0.1:8001/apis/custom.metrics.k8s.io/v1beta1
	kubectl get --raw /apis/external.metrics.k8s.io/v1beta1 # http://127.0.0.1:8001/apis/external.metrics.k8s.io/v1beta1

tls:
	openssl req -new -newkey rsa:4096 -x509 -sha256 -days 365 -nodes -out serving.crt -keyout serving.key -subj "/C=CN/CN=custom-metrics-apiserver.kube-system.svc.cluster.local"
