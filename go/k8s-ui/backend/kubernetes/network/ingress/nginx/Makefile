
USER := $(shell id -u)
GROUP := $(shell id -g)
POD_NAME := $(shell kubectl get pods -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -o=jsonpath='{.items[0].metadata.name}')
POD_NAMESPACE=ingress-nginx

pc:
	docker container rm openresty-pc
	docker run -p 8088:80 --name openresty-pc \
	-u $(USER):$(GROUP) \
	-v $(PWD)/index.html:/var/www/html/pc/index.html \
	-v $(PWD)/openresty-default.conf:/etc/openresty/nginx.conf \
	openresty/openresty:1.15.8.3-1-centos

nginx:
	kubectl exec -it $(POD_NAME) -n $(POD_NAMESPACE) -- cat /etc/nginx/nginx.conf > ingress-nginx-raw.conf

sh:
	kubectl exec -it $(POD_NAME) -n $(POD_NAMESPACE) -- /bin/sh

curl:
	# curl "$(minikube service web --url)"
	kubectl exec -it $(POD_NAME) -n $(POD_NAMESPACE) -- curl ingress-nginx:8080

version:
	kubectl exec -it $(POD_NAME) -n $(POD_NAMESPACE) -- /nginx-ingress-controller --version
