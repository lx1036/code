
# Install & Upgrade
安装文档：https://cert-manager.io/docs/installation/kubernetes/
github: https://github.com/jetstack/cert-manager

```shell script
kubectl create namespace cert-manager
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.13.1/cert-manager.yaml
kubectl get pods -n cert-manager --watch

kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/${version}/cert-manager.yaml # upgrade
kubectl delete -f https://github.com/jetstack/cert-manager/releases/download/${version}/cert-manager.yaml # uninstall
```

```shell script
cat <<EOF > test-resources.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: cert-manager-test
---
apiVersion: cert-manager.io/v1alpha2
kind: Issuer
metadata:
  name: test-selfsigned
  namespace: cert-manager-test
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: selfsigned-cert
  namespace: cert-manager-test
spec:
  dnsNames:
    - example.com
  secretName: selfsigned-cert-tls
  issuerRef:
    name: test-selfsigned
EOF
kubectl apply -f test-resources.yaml
```
