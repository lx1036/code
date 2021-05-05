
# CKA Curriculum 考试大纲 
https://github.com/cncf/curriculum/blob/master/CKA_Curriculum_v1.20.pdf

CKA 考试内容(总共20道题，2个小时):
集群架构，安装和配置：25%(5道题)
* 管理基于角色的访问控制（RBAC） Manage role based access control (RBAC)
* 使用Kubeadm安装基本集群 Use Kubeadm to install a basic cluster
* 管理高可用性的Kubernetes集群 Manage a highly-available Kubernetes cluster
* 设置基础架构以部署Kubernetes集群 Provision underlying infrastructure to deploy a Kubernetes cluster
* 使用Kubeadm在Kubernetes集群上执行版本升级 Perform a version upgrade on a Kubernetes cluster using Kubeadm
* 实施etcd备份和还原 Implement etcd backup and restore

工作负载和调度：15%(3道题)
* 了解部署以及如何执行滚动更新和回滚 Understand deployments and how to perform rolling update and rollbacks
* 使用ConfigMaps和Secrets配置应用程序 Use ConfigMaps and Secrets to configure applications
* 了解如何扩展应用程序 Know how to scale applications
* 了解用于创建健壮的、自修复的应用程序部署的原语 Understand the primitives used to create robust, self-healing, application deployments
* 了解资源限制如何影响Pod调度 Understand how resource limits can affect Pod scheduling
* 了解清单管理和通用模板工具 Awareness of manifest management and common templating tools

服务和网络：20%(4道题)
* 了解集群节点上的主机网络配置 Understand host networking configuration on the cluster nodes
* 理解Pods之间的连通性 Understand connectivity between Pods
* 了解ClusterIP、NodePort、LoadBalancer服务类型和端点 Understand ClusterIP, NodePort, LoadBalancer service types and endpoints
* 了解如何使用入口控制器和入口资源 Know how to use Ingress controllers and Ingress resources
* 了解如何配置和使用CoreDNS Know how to configure and use CoreDNS
* 选择适当的容器网络接口插件 Choose an appropriate container network interface plugin

存储：10%(2道题)
* 了解存储类、持久卷 Understand storage classes, persistent volumes
* 了解卷模式、访问模式和卷回收策略 Understand volume mode, access modes and reclaim policies for volumes
* 理解持久容量声明原语 Understand persistent volume claims primitive
* 了解如何配置具有持久性存储的应用程序 Know how to configure applications with persistent storage

故障排除：30%(6道题)
* 评估集群和节点日志 Evaluate cluster and node logging
* 了解如何监视应用程序 Understand how to monitor applications
* 管理容器标准输出和标准错误日志 Manage container stdout & stderr logs
* 解决应用程序故障 Troubleshoot application failure
* 对群集组件故障进行故障排除 Troubleshoot cluster component failure
* 排除网络故障 Troubleshoot networking


# CKA 真题
2020-11 和 2019-07 真题
模拟题：https://rx-m.com/cka-online-training/

```shell
# 切换 k8s 集群
kubectl config use-context k8s
```

## Module 1 - Cluster Architecture, Installation, and Configuration
https://rx-m.com/cka-online-training/ckav2-online-training-module-1/

(1)RBAC: 创建一个 deployment-clusterrole ClusterRole，只具有创建 "deployments", "statefulsets", "daemonset" 资源的权限，
并在 Namespace app-team1 创建 cicd-token ServiceAccount，并把 cicd-token ServiceAccount 绑定到 deployment-clusterrole ClusterRole 上。
https://blog.csdn.net/shenhonglei1234/article/details/109413090
```shell
kubectl create namespace app-team1
kubectl create clusterrole deployment-clusterrole --verb=create --resource=deployments,statefulsets,daemonsets
kubectl create serviceaccount cicd-token -n app-team1
kubectl create clusterrolebinding deployment-clusterrolebinding --clusterrole=deployment-clusterrole --serviceaccount=app-team1:cicd-token
```

(2)升级集群: 将集群中 master 所有组件从 v1.18 升级到 1.19(controller,apiserver,scheduler,kubelet,kubectl)？
参考：https://kubernetes.io/zh/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/
[我的k8s升级原则]控制组件kube-apiserver/kube-controller-manager/kube-scheduler版本保持一致；计算组件kubelet/kube-proxy版本保持一致，且必须比控制组件小一个版本。
```shell
# `kubeadm upgrade apply v1.19.0` 命令不会升级 kubelet，需要手动升级
kubectl cordon k8s-master
kubectl drain k8s-master --ignore-daemonsets --force
apt-get install kubeadm=1.19.0-00 kubelet=1.19.0-00 kubectl=1.19.0-00
systemctl daemon-reload && systemctl restart kubelet
kubeadm upgrade apply v1.19.0
```

## Module 2 - Workloads and Scheduling
https://rx-m.com/cka-online-training/ckav2-online-training-module-2/
(1)scale: 将一个 Deployment 的副本数量从 1 个副本扩至3 个？
```shell
kubectl scale --current-replicas=1 --replicas=3 deployment/nginx
```

(2)Pod: 创建一个pod，包含多个image，如 image=nginx,name=nginx; image=redis,name=redis ？
```shell
kubectl create deployment test-deploy --image=nginx:1.17.8 --port=80
kubectl edit deploy test-deploy # 手动添加多个容器
```

(3)Schedule: 将名为 ek8s-node-1 的 node 设置为不可用，并重新调度该 node 上所有 运行的 pods
```yaml
kubectl cordon ek8s-node-1
kubectl drain ek8s-node-1 --ignore-daemonsets --force
```

(4)查询集群中节点，找出可以调度节点的数量，(其实就是被标记为不可调度和 打了污点的节点之外的节点 数量 )，然后将数量写到指定文件？
检查有多少 worker nodes 已准备就绪(不包括被打上 Taint:NoSchedule 的节点)， 并将数量写入 /opt/KUSC00402/kusc00402.txt
```shell
# 查询集群 Ready 节点数量
kubectl get node | grep -i ready
# 判断节点有无不可调度污点
kubectl describe nodes <nodeName> | grep -i taints | grep -i noSchedule
```

(5)Tolerations: 确保在 kubectl 集群的每个节点上运行一个 Nginx Pod。其中 Nginx Pod 必须使用 Nginx 镜像。不要覆盖当前环境中的任何 taints。 使用 Daemonset 来完成这个任务，Daemonset 的名字使用 ds。
```yaml
apiVersion: apps/v1
	kind: DaemonSet
	metadata:
	  name: ds
	  namespace: kube-system
	  labels:
		k8s-app: fluentd-logging
	spec:
	  tolerations:
			# this toleration is to have the daemonset runnable on master nodes
			# remove it if your masters can't run pods
			- key: node-role.kubernetes.io/master
			  effect: NoSchedule
	  selector:
		matchLabels:
		  name: fluentd-elasticsearch
	  template:
		metadata:
		  labels:
			name: fluentd-elasticsearch
		spec:
		  containers:
		  - name: fluentd-elasticsearch
			image: nginx
```

(6)initContainers: 添加一个 initcontainer 到 lum(/etc/data)这个 initcontainer 应该创建一个名为/workdir/calm.txt 的空文件，如果/workdir/calm.txt 没有被检测到，这个 Pod 应该退出
```yaml
      #题目中yaml文件已经给出，只需要增加initcontainers部分，以及emptyDir: {} 即可
      #init文档位置：https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    env: test
spec:
  volumes:
  - name: workdir
    emptyDir: {} 
  containers:
  - name: nginx
    image: nginx
    command: [if ..]
    volumeMounts:
      - name: work
        mountPath: /workdir
  initContainers:
  - name: init-myservice
    image: busybox:1.28
    command: ["sh", "-c", "touch /workdir/calm.txt"] 
    volumeMounts:
    - name: work
      mountPath: /workdir
```

(7)Deployment: 创建 deployment 名字为 nginx-app 容器采用 1.11.9 版本的 nginx  这个 deployment 包含 3 个副本,接下来通过滚动升级的方式更新镜像版本为 1.12.0，并记录这个更新，最后，回滚这个更新到之前的 1.11.9 版本
创建 deployment 的 spec 文件: 使用 redis 镜像，7 个副本，label 为 app_enb_stage=dev deployment 名字为 kual00201 保存这个 spec 文件到/opt/KUAL00201/deploy_spec.yaml完成后，清理(删除)在此任务期间生成的任何新的 k8s API 对象
```shell
kubectl run deployment nginx-app --image=nginx:1.11.9 --replicas=3
kubectl set image deployment nginx-app nginx-app=nginx:1.12.0 --record  (nginx-app container名字)
kubectl rollout history deployment nginx-app
kubectl rollout undo deployment nginx-app

kubectl create deploy kual00201 --image=redis --labels=app_enb_stage=dev --dry-run -o yaml > /opt/KUAL00201/deploy_spec.yaml
```

## Module 3 - Services and Networking
https://rx-m.com/cka-online-training/ckav2-online-training-module-3/
(1)Ingress: 创建 Ingress，将指定的 Service 的 9999 端口在/test 路径暴露出来？
```shell
# foo.com 域名证书在 my-cert secret 里
kubectl create ingress test-ingress --rule="foo.com/bar=svc1:8080,tls=my-cert"
```

INFO: NetworkPolicy 需要重新演练几遍
(2)NetworkPolicy: 在指定namespace创建一个NetworkPolicy, 允许namespace中的Pod访问同namespace中其他Pod的8080端口？
创建一个名为 allow-port-from-namespace 的新 NetworkPolicy，以允许现有 namespace corp-net 中的 Pods 连接到同一 namespace 中其他 Pods 的端 口 9200。
确保新的 NetworkPolicy:不允许对没有在监听端口 9200 的 Pods 的访问；不允许不来自 namespacecorp-net 的 Pods 的访问
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
  namespace: default
spec:
  podSelector:
    matchLabels:
      app: backend
  policyTypes:
  - Egress
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: db
      ports:
        - port: 8080
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: all-port-from-namespace
  namespace: internal
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector: 
       matchLabels: 
         name: namespacecorp-net
	- podSelector: {}
	  ports:
	  - port: 9200
```

(3)将集群中一个 Deployment 服务暴露出来(是一个 nginx，使用kubectl expose 命令暴露即可)？
请重新配置现有的部署 front-end 以及添加名为 http 的端口规范来公开现 有容器 nginx 的端口 80/tcp。
创建一个名为 front-end-svc 的新服务，以公开容器端口 http。 配置此服务，以 通过在排定的节点上的 NodePort 来公开各个 Pods 考点:将现有的 deploy 暴露成 nodeport 的 service。
```shell
kubectl create deployment front-end --image=nginx --port=80
kubectl expose deployment front-end --port=80 --target-port=80 --name=front-end-svc --type=NodePort
```

(4)Service: 创建一个文件/opt/kucc.txt ，这个文件列出所有的 service 为 foo ,在 namespace 为 production 的 Pod这个文件的格式是每行一个 Pod的名字
```shell
kubectl get svc -n production --show-labels | grep foo
kubectl get pods -l app=foo -o=custom-columns=NAME:.spec.name > kucc.txt
```

(5)创建一个 deployment 名字为:nginx-dns 路由服务名为：nginx-dns 确保服务和 pod 可以通过各自的 DNS 记录访问 容器使用 nginx 镜像，
使用 nslookup 工具来解析 service 和 pod 的记录并写入相应的/opt/service.dns 和/opt/pod.dns 文件中，确保你使用 busybox:1.28 的镜像用来测试。
```
    busybox这里找：https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/
    1.kubectl run nginx-dns --image=nginx
      kubectl expose deployment nginx-dns --name=nginx-dns --port=80 --type=NodePort
      kubectl get pod -owide xxx (查看pod IP)
    2.建立busybox
       apiVersion: v1
    	kind: Pod
    	metadata:
      name: busybox1
      labels:
        name: busybox
    	spec:
      hostname: busybox-1
      subdomain: default-subdomain
      containers:
    - image: busybox:1.28
      command:
      sleep
      "3600"
      name: busybo
    3.解析
      kubectl exec -it busybox -- nslookup nginx-dns
      kubectl exec -it busybox -- nslookup 10.244.0.122(pod IP)
```


## Module 4 - Storage
https://rx-m.com/cka-online-training/ckav2-online-training-module-4/
(1)Etcd: 对 etcd 进行 snapshot save 和 restore，因为 https 会提供 endpoints, cacert, cert 和 key？
```shell
# etcdctl 3.4.10 以上好像不需要指定 api 版本了，已经默认了
ETCDCTL_API=3 etcdctl --endpoints="https://127.0.0.1:2379" --cacert=ca.crt --cert=etcd.crt --key=etcd.key snapshot save /etc/data/etcd-snapshot.db
ETCDCTL_API=3 etcdctl --endpoints="https://127.0.0.1:12379" --cacert=ca.crt --cert=etcd.crt --key=etcd.key snapshot restore /etc/data/etcd-snapshot.db
```

(2)PVC/PV: 对集群中的 PV 按照大小顺序排序显示，并将结果写到指定文件？
```shell
kubectl get pv --sort-by=.spec.capacity.storage --no-headers > pv.txt
```

(3)PVC/PV: 创建一个 Name: web-server, Image: nginx, mountPath: /usr/share/nginx/html, 同时 volume 具有 ReadWriteOnce 权限？
使用指定 storageclass 创建一个 pvc，大小为 10M，将这个 nginx 容器的/var/nginx/html目录使用该 pvc 挂在出来，将这个 pvc 的大小从 10M 更新成 70M?
```yaml
#通过 kubectl edit pvc pv-volume 可以进行修改容量
apiVersion: v1
kind: Pod
metadata:
  name: web-server
spec:
  containers:
    - name: nginx-1
      image: nginx
      volumeMounts:
        - name: data
          mountPath: /usr/share/nginx/html
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: hostpath-pvc
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: hostpath-pvc
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  resources:
    requests:
      storage: 8Gi
  selector:
    matchLabels:
      pv: hostpath
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: hostpath-pv
  labels:
    pv: hostpath
spec:
  capacity:
    storage: 100Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: /usr/share/nginx/html
```

(5)Secret: 创建一个secret,名字为super-secret包含用户名bob,创建pod1挂载该secret，路径为/secret，创建pod2，使用环境变量引用该secret，该变量的环境变量名为ABC
```yaml
#https://kubernetes.io/zh/docs/concepts/configuration/secret/#%E8%AF%A6%E7%BB%86
#	echo -n "bob" | base64
	
apiVersion: v1
kind: Secret
metadata:
  name: super-secret
type: Opaque
data:
  username: Ym9i
---
apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  containers:
  - name: mypod
    image: redis
    volumeMounts:
    - name: foo
      mountPath: "/secret"
      readOnly: true
  volumes: secret
  - name: foo
    secret:
      secretName: super-secret
---
apiVersion: v1
kind: Pod
metadata:
  name: pod-evn-eee
spec:
  containers:
  - name: mycontainer
    image: redis
    env:
      - name: ABC
        valueFrom:
          secretKeyRef:
            name: super-secret
            key: username
```

## Module 5 - Troubleshooting
https://rx-m.com/cka-online-training/ckav2-online-training-module-5/
(1)列出指定pod的日志中状态为Error的行，并记录在指定的文件上? 同时，集群中存在一个 Pod，并且该 Pod 中的容器会将 log 输出到指定文件。
修改 Pod 配置，将 Pod 的日志输出到控制台,其实就是给 Pod 添加一个 sidecar，然后不断读取指定文件，输出到控制台？
```shell
kubectl logs deployment/nginx -c nginx-1 | grep "Error" > /opt/KUCC000xxx/KUCC000xxx.txt
```
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: podname
spec:
  containers:
  - name: count
    image: busybox
    args: ["/bin/sh", "-c", "i=0;while true;do echo '$(date) INFO $i' >> /var/log/legacy-ap.log;i=$((i+1));sleep 1;done"]
    volumeMounts:
    - name: logs
      mountPath: /var/log
  - name: count-log-1
    image: busybox
    args: ["/bin/sh", "-c", "tail -n+1 -f /var/log/legacy-ap.log"]
    volumeMounts:
    - name: logs
      mountPath: /var/log
  volumes:
  - name: logs
    emptyDir: {}

# 验证:
# kubectl logs <pod_name> -c <container_name>
```

(2)找集群中带有指定 label 的 Pod 中占用资源最高的，并将它的名字写入指定的文件？
通过 pod label name=cpu-loader，找到运行时占用大量 CPU 的 pod， 并将占用 CPU 最高的 pod 名称写入文件 /opt/KUTR000401/KUTR00401.txt(已 存在)。
```shell
kubectl top pod -l name=cpu-user -A
#NAMAESPACE NAME CPU MEM
#delault cpu-user-1 45m 6Mi
#delault cpu-user-2 38m 6Mi
#delault cpu-user-3 35m 7Mi
#delault cpu-user-4 32m 10Mi
echo 'cpu-user-1' >>/opt/KUTR00401/KUTR00401.txt
```

(3)static pod: 题目很绕，大致是 在k8s的集群中的node1节点配置kubelet的service服务，去拉起一个由kubelet直接管理的pod(说明了是静态pod)，
```
该文件应该放置在/etc/kubernetes/manifest目录下(给出了pod路径)

创建  1.vi /etc/kubernetes/manifest/static-pod.yaml
      2.systemctl status kubelet   查找kubelet.service路径  考试目录是: /etc/systemd/system/kubernetes.service
	  3.vi /etc/systemd/system/kubernetes.service   观察有没有 --pod-manifest-path=/etc/kubernetes/manifest 这句话   没有就加上得
	  4.sudo -i   ssh node  sudo -i
	  5.systemctl daemon-reload            systemctl restart kubelet.service
	  6.systemctl enable kubelet
      7.检查  kubectl get pods -n kube-system | grep static-pod  实际是static-pod+系统  static-pod-kubelet-service
```
