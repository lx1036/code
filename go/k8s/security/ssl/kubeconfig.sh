
# 查看证书信息
openssl x509 -noout -text -in ca.crt

openssl genrsa -out developer.key 2048 # 创建证书私钥
openssl req -new -key developer.key -out developer.csr -subj "/CN=developer" # 私钥创建一个 csr(证书签名请求)文件
openssl x509 -req -in developer.csr -CA ~/.minikube/ca.crt -CAkey ~/.minikube/ca.key -CAcreateserial -out developer.crt -days 365 # 为用户颁发证书

kubectl create clusterrolebinding kubernetes-viewer --clusterrole=view --user=developer # 绑定view这个cluster role到developer用户

# 基于客户端证书生成 Kubeconfig 文件 kubeconfig.yml，验证只有读权限，没有写权限
kubectl --kubeconfig=kubeconfig.yml --context=minikube delete pods nginx-demo-2-8d544cc7-pm5qk
kubectl --kubeconfig=kubeconfig.yml --context=minikube delete pods nginx-demo-2-8d544cc7-pm5qk
kubectl --kubeconfig=kubeconfig.yml --context=minikube apply -f ../../nginx/minikube-nginx.yml
