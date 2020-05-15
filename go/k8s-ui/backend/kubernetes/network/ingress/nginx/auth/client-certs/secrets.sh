
cat ca.crt server.crt client.crt >> ca-2.crt
kubectl create secret generic ca-secret --from-file=ca.crt=ca-2.crt
kubectl create secret generic tls-secret --from-file=tls.crt=server.crt --from-file=tls.key=server.key
