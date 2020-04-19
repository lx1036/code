
#curl "$(minikube service web --url)"

kubectl exec -it \
  "$(kubectl get pods -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -o=jsonpath='{.items[0].metadata.name}')" \
  -n ingress-nginx -- cat /etc/nginx/nginx.conf > ingress-nginx-raw.conf

# 这是可以访问的
kubectl exec -it \
  "$(kubectl get pods -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -o=jsonpath='{.items[0].metadata.name}')" \
  -n ingress-nginx -- curl ingress-nginx:8080
