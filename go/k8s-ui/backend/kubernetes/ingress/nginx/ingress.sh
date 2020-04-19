
curl "$(minikube service web --url)"

kubectl exec -it \
  "$(kubectl get pods -n kube-system -l app.kubernetes.io/name=nginx-ingress-controller -o=jsonpath='{.items[0].metadata.name}')" \
  -n kube-system -- cat /etc/nginx/nginx.conf > ingress-nginx-raw.conf
