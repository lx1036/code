
#cat ca.crt server.crt client.crt >> ca-2.crt
#kubectl create secret generic ca-secret --from-file=ca.crt=ca.crt
#kubectl create secret generic tls-secret --from-file=tls.crt=server.crt --from-file=tls.key=server.key



kubectl delete secret ca-secret
kubectl create secret generic ca-secret --from-file=tls.crt=server.crt --from-file=tls.key=server.key --from-file=ca.crt=ca.crt
