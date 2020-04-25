
eval $(minikube docker-env)
curl $(minikube ip):2020/status
open http://$(minikube ip):2021/metrics
docker container logs openresty

