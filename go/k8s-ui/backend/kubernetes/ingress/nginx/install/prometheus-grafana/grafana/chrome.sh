

PORT=$(kubectl get svc grafana-server -o=jsonpath='{.spec.ports[0].nodePort}' -n ingress-nginx)
IP=$(minikube ip)
open http://"${IP}":"${PORT}"


