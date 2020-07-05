

**[ca](https://github.com/kubernetes/ingress-nginx/blob/master/docs/examples/PREREQUISITES.md#client-certificate-authentication)**
**[client-certs](https://github.com/kubernetes/ingress-nginx/blob/master/docs/examples/auth/client-certs/README.md)**

# ssl - Kubernetes将ca证书添加到pods的trust root中
```shell script
COPY my-cert.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates
```
