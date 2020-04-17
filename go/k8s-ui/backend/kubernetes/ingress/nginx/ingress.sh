
curl "$(minikube service web --url)"

kubectl exec -it nginx-ingress-controller-6d57c87cb9-hmj2c -n kube-system -- cat /etc/nginx/nginx.conf > ingress-nginx-raw.conf
