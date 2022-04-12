
# LoadBalancer Service IPAM Controller
(1)支持 multi-ippool，可以根据 service annotation 来选择对应的 ippool，默认选择 default ippool
```yaml
---

apiVersion: bgplb.k9s.io/v1
kind: IPPool
metadata:
  name: default
spec:
  cidr: "100.20.30.0/24"

---

apiVersion: bgplb.k9s.io/v1
kind: IPPool
metadata:
  name: ippool2
spec:
  cidr: "200.20.30.0/24"

```

(2)ipam controller 重启后，重新 restore from 已有的 LoadBalancer Service IP，不会为新建的 LoadBalancer Service 重复分配 IP

(3)可以动态添加 ippool，而不需要重启 ipam controller pod
```yaml

---

apiVersion: v1
kind: Service
metadata:
  name: nginx-custom-ippool
  annotations:
    loadbalancer/ippool-name: "ippool2"
spec:
  selector:
    app: nginx-demo
  ports:
    - port: 80
  type: LoadBalancer

---
```
