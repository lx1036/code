

PORT=$(kubectl get svc grafana -o=jsonpath='{.spec.ports[0].nodePort}' -n monitoring)
IP=$(minikube ip)
open http://"${IP}":"${PORT}"


