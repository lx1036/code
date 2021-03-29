

kubectl create secret tls metrics-server --cert=./metrics-server.pem --key=./metrics-server-key.pem -n kube-system

