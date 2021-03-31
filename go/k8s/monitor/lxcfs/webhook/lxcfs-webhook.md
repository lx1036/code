



## 证书生成
一键生成 HTTPS server.crt/server.key 证书文件，自己作为 CA 签名

```shell
openssl req -new -newkey rsa:4096 -x509 -sha256 -days 365 -nodes -out serving.crt -keyout serving.key -subj "/C=CN/CN=custom-metrics-apiserver.cattle-prometheus.svc.cluster.local"
# And you will find serving.crt and serving.key in your path. And then you are going to create a secret in cattle-prometheus namespace.
kubectl create secret generic -n cattle-prometheus cm-adapter-serving-certs --from-file=serving.key=./serving.key --from-file=serving.crt=./serving.crt 

```


## 参考文献
https://github.com/denverdino/lxcfs-admission-webhook

https://github.com/xigang/lxcfs-admission-webhook


https://xigang.github.io/2019/11/09/lxcfs-admission-webhook/
