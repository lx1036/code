

PORT=$(kubectl get svc prometheus-server -o=jsonpath='{.spec.ports[0].nodePort}' -n ingress-nginx)
IP=$(minikube ip)
open http://"${IP}":"${PORT}"


