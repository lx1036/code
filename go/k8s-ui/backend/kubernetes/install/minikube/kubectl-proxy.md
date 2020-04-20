
```shell script
# 这样可以直接在局域网内任何一台服务器上 curl http://192.168.31.35:8001
kubectl proxy --address='0.0.0.0' --disable-filter=true

# curl --cacert ~/.minikube/ca.crt --cert ~/.minikube/profiles/minikube/client.crt --key ~/.minikube/profiles/minikube/client.key http://localhost:8001
kubectl proxy
```
