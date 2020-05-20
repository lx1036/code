


kubectl run nginx --image=nginx
kubectl get pods -l run=nginx



# Validation

POD_NAME=$(kubectl get pods -l run=nginx -o jsonpath="{.items[0].metadata.name}")
kubectl port-forward "$POD_NAME" 6789:80
# port-forward
curl --head localhost:6789
# logs
kubectl logs "$POD_NAME"
kubectl exec -ti "$POD_NAME" -- nginx -v
