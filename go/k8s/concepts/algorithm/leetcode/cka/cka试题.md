

(1) https://blog.csdn.net/shenhonglei1234/article/details/109413090
Create a new ClusterRole named deployment-clusterrole that only allows the creation of the following resource types:
Deployment
StatefulSet
DaemonSet
Create a new ServiceAccount named cicd-token in the existing namespace app-team1.
Limited to namespace app-team1, bind the new ClusterRole deployment-clusterrole to the new ServiceAccount cicd-token.

```yaml
# 参考文档：
# https://kubernetes.io/zh/docs/reference/access-authn-authz/rbac/#kubectl-create-clusterrolebinding
# https://kubernetes.io/zh/docs/reference/access-authn-authz/service-accounts-admin/
---
# kubectl create namespace app-team1
# kubectl delete namespace app-team1
apiVersion: v1
kind: Namespace
metadata:
  name: app-team1
  
---

# kubectl create clusterrole deployment-clusterrole --verb=create --resource=deployments,statefulsets,daemonsets
# kubectl delete clusterrole deployment-clusterrole
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: deployment-clusterrole
rules:
  - verbs: ["create"]
    apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "daemonset"]

---
# kubectl create serviceaccount cicd-token -n app-team1
# kubectl delete serviceaccount cicd-token -n app-team1
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cicd-token
  namespace: app-team1

---
# kubectl create clusterrolebinding deployment-clusterrolebinding --clusterrole=deployment-clusterrole --serviceaccount=app-team1:cicd-token
# kubectl delete clusterrolebinding deployment-clusterrolebinding
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: deployment-clusterrolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: deployment-clusterrole
subjects:
  - kind: ServiceAccount
    name: cicd-token
    namespace: app-team1

```


(2) https://blog.csdn.net/shenhonglei1234/article/details/109413090
Set the node named ek8s-node-1 as unavaliable and reschedule all the pods running on it.


# CKA Curriculum: https://github.com/cncf/curriculum/blob/master/CKA_Curriculum_v1.19.pdf


## Module 1 - Cluster Architecture, Installation, and Configuration
https://rx-m.com/cka-online-training/ckav2-online-training-module-1/
第一部分主要考察内容：
* RBAC
* [我的k8s升级原则]控制组件kube-apiserver/kube-controller-manager/kube-scheduler版本保持一致；计算组件kubelet/kube-proxy版本保持一致，且必须比控制组件小一个版本。






## Module 2 - Workloads and Scheduling
https://rx-m.com/cka-online-training/ckav2-online-training-module-2/



## Module 3 - Services and Networking
https://rx-m.com/cka-online-training/ckav2-online-training-module-3/




## Module 4 - Storage
https://rx-m.com/cka-online-training/ckav2-online-training-module-4/




## Module 5 - Troubleshooting
https://rx-m.com/cka-online-training/ckav2-online-training-module-5/



# CKA 2020-11~12 真题！！！
(1)创建 clusterrole,并且对该 clusterrole 只绑定对 Deployment， Daemonset,Statefulset 的创建权限；
在指定 namespace 创建一个 serviceaccount，并且将上一步创建 clusterrole 和该 serviceaccount 绑定？
```yaml
#解答
#创建对应的 ClusterRole,并绑定对应的权限
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
name: deployment-clusterrole
rules:
- apiGroups: ["apps"]
resources: ["daemonsets", "deployments", "statefulsets"]
verbs: ["create"]

#创建对应的 serviceaccount
  kubectl -n app-team1 create serviceaccount cicd-token

#将 serviceaccount 与 ClusterRole 进行绑定
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
name: read-secrets-global
subjects:
  - kind: ServiceAccount
name: cicd-token
namespace: app-team1
roleRef:
kind: ClusterRole
name: deployment-clusterrole
apiGroup: rbac.authorization.k8s.io
```
1、Context k8s
为部署管道创建一个新的 ClusterRole 并将其绑定到范围为特定的 namespace
的特定 ServiceAccount。 Task
创建一个名为 deployment-clusterrole 且仅允许创建以下资源类型的新 ClusterRole:
Deployment StatefulSet DaemonSet
在现有的 namespace app-team1 中创建一个名为 cicd-token 的新 ServiceAccount。限于 namespace app-team1，将新的 ClusterRole deployment-clusterrole 绑定到新的 ServiceAccount cicd-token
考点:RABC 授权模型的理解。
kubectl create clusterrole deployment-clusterrole --verb=create --
resource=deployments,statefulsets,daemonsets
kubectl create serviceaccount cicd-token --namespace=default
kubectl create rolebinding deployment-clusterrole --clusterrole=deployment- clusterrole --serviceaccount=default:cicd-token --namespace=default



(2)对指定 etcd 集群进行备份和还原,考试时会给定 endpoints, 根证书，证书签名，私钥？
```shell
#解答
首先，为运行在 https://127.0.0.1:2379 上的现有 etcd 实力创建快照并且将快照保存到
/etc/data/etcd-snapshot.db
然后还原与/var/lib/backup/etcd-snapshot-previoys.db 的现有先前快照
提供了以下 TLS 证书和密钥，已通过 etcdctl 连接到服务器
ca 证书:/opt/KUIN000601/ca.crt
客户端证书:/opt/KUIN000601/etcd-client.crt
客户端密钥:/opt/KUIN000601/etcd-client.key
#备份:要求备份到指定路径及指定文件名
$ ETCDCTL_API=3 etcdctl --endpoints="https://127.0.0.1:2379" --
cacert=/opt/KUIN000601/ca.crt --cert=/opt/KUIN000601/etcd-client.crt --
key=/opt/KUIN000601/etcd-client.key snapshot save /etc/data/etcd-snapshot.db
#还原:要求使用指定文件进行还原
$ ETCDCTL_API=3 etcdctl --endpoints="https://127.0.0.1:2379" --
cacert=/opt/KUIN000601/ca.crt --cert=/opt/KUIN000601/etcd-client.crt --
key=/opt/KUIN000601/etcd-client.key snapshot restore /var/lib/backup/etcd-snapshot-previoys.db
```

4、此项目无需更改配置环境。问题权重: 7%
Task
首先，为运行在 https://127.0.0.1:2379 上的现有 etcd 实例创建快照并将快照 保存到 /srv/data/etcd-snapshot.db。
为给定实例创建快照预计能在几秒钟内完成。 如果该操作似乎挂起，则命令可 能有问题。用 + 来取消操作，然后重试。
然后还原位于/data/backup/etcd-snapshot-previous.db 的现有先前快照。 提供了以下 TLS 证书和密钥，以通过 etcdctl 连接到服务器。
CA 证书: /opt/KUIN00601/ca.crt
客户端证书: /opt/KUIN00601/etcd-client.crt
客户端密钥: /opt/KUIN00601/etcd-client.key 考点:etcd 的备份和还原命令
ETCDCTL_API=3 etcdctl --endpoints $ENDPOINT snapshot save/restore snapshotdb --cert=/opt/KUIN00601/etcd-client.crt --key=/opt/KUIN00601/etcd- client.key --cacert=/opt/KUIN00601/ca.crt




(3)升级集群，将集群中 master 所有组件从 v1.18 升级到 1.19(controller,apiserver,scheduler,kubelet,kubectl)？
```shell
#解答
#将节点标记为不可调度状态
$ kubectl cordon k8s-master
#驱逐节点上面的 pod
$ kubectl drain k8s-master--delete-local-data --ignore-daemonsets --force
#升级组件
$ apt-get install kubeadm=1.19.0-00 kubelet=1.19.0-00 kubectl=1.19.0-00
#重启 kubelet 服务
$ systemctl restart kubelet
#升级集群其他组件
$ kubeadm upgrade apply v1.19.0
```

2、将名为 ek8s-node-1 的 node 设置为不可用，并重新调度该 node 上所有 运行的 pods
考点:cordon 和 drain 命令的使用
$ kubectl cordon ek8s-node-1
$ kubectl drain ek8s-node-1 --force

3、现有的 Kubernetes 集群正在运行版本 1.18.8。仅将主节点上的所有 Kubernetes 控制平面和节点组件升级到版本 1.19.0。
另外，在主节点上升级 kubelet 和 kubectl。
确保在升级之前 drain 主节点，并在升级后 uncordon 主节点。 请不要升级 工作节点，etcd，container 管理器，CNI 插件， DNS 服务或任何其他插件。

考点:如何离线主机，并升级控制面板和升级节点
kubectl drain <cp-node-name> --ignore-daemonsets
sudo kubeadm upgrade apply v1.19.0
yum install -y kubelet-1.19.0 kubectl-1.19.0 --disableexcludes=kubernetes
sudo systemctl daemon-reload
sudo systemctl restart kubelet
kubectl uncordon <cp-node-name>
--升级节点
sudo kubeadm upgrade node
yum install -y kubelet-1.19.0 kubectl-1.19.0 --disableexcludes=kubernetes
sudo systemctl daemon-reload
sudo systemctl restart kubelet


(4)创建 Ingress，将指定的 Service 的 9999 端口在/test 路径暴露出来？
```shell
#解答
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
name: pong
namespace: ing-internal
annotations:
nginx.ingress.kubernetes.io/rewrite-target: /
spec:
rules:
- http:
paths:
- path: /hi
pathType: Prefix
backend:
service:
name: hi
port:
number: 5678
```

7、问题权重: 7%设置配置环境: kubectl config use-context k8s
Task
如下创建一个新的 nginx Ingress 资源: 名称: ping
Namespace: ing-internal
使用服务端口 5678 在路径 /hello 上公开服务 hello 可以使用以下命令检查服务 hello 的可用性，该命令应返回 hello: curl -kL <INTERNAL_IP>/hello
考点:Ingress 的创建
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
name: ping
namespace: ing-internal
annotations:
nginx.ingress.kubernetes.io/rewrite-target: /
spec:
rules:
- http:
  paths:
- path: /hello
  pathType: Prefix
  backend:
  service:
  name: hello
  port:
  number: 5678  



(5)创建一个两个 container 容器的 Pod:nginx+redis？

11、设置配置环境:问题权重: 4%
kubectl config use-context k8s
Task
创建一个名为 kucc8 的 pod，在 pod 里面分别为以下每个 images 单独运 行一个 app container(可能会有 1-4 个 images):nginx + redis + memcached + consul
考点:pod 概念 apiVersion: v1 kind: Pod metadata: name: kucc1 spec: containers:
- image: nginx name: nginx
- image: redis name: redis
- image: memchached
  name: memcached
- image: consul
  name: consul




(6)对集群中的 PV 按照大小顺序排序显示，并将结果写道指定文件？



(7)将一个 Deployment 的副本数量从 1 个副本扩至3 个？
8、设置配置环境:问题权重: 4%
kubectl config use-context k8s
Task
将 deployment 从 presentation 扩展至 6 pods 考点:kubectl scale 命令
$ kubectl scale --replicas=6 deployment/loadbalancer


(8)在指定 namespace 创建一个 Networkpolicy, 允许 namespace 中的 Pod 访问 同 namespace 中其他 Pod 的8080 端口？



(9)集群中存在一个 Pod，并且该 Pod 中的容器会将 log 输出到指定文件。 修改 Pod 配置，将 Pod 的日志输出到控制台,其实就是给 Pod 添加一个 sidecar，然后不断读取指定文件，输出到控制台？
```yaml
apiVersion: v1
kind: Pod
metadata:
name: podname
spec:
containers:
- name: count
image: busybox
args:
- /bin/sh
- -c
->
i=0;
while true;
do
echo "$(date) INFO $i" >> /var/log/legacy-ap.log;
i=$((i+1));
sleep 1;
done
volumeMounts:
- name: logs
mountPath: /var/log
- name: count-log-1
image: busybox
args: [/bin/sh, -c, 'tail -n+1 -f /var/log/legacy-ap.log']
volumeMounts:
  - name: varlog
mountPath: /var/log
volumes:
  - name: logs
emptyDir: {}

# 验证:
# kubectl logs <pod_name> -c <container_name>
```

16、 设置配置环境:问题权重: 7%
kubectl config use-context k8s
在不更改其现有容器的情况下，需要将一个现有的 Pod 集成到 Kubernetes 的内置日志记录体系结构中(例如 kubectl logs)。添加 streaming sidecar 容器是实现此要求的一种好方法。
Task
将一个 busybox sidecar 容器添加到现有的 Pod legacy-app。新的 sidecar
容器必须运行以下命令:
/bin/sh -c tail -n+1 -f /var/log/legacy-app.log
使用名为 logs 的 volume mount 来让文件 /var/log/legacy-app.log 可用于 sidecar 容器。
不要更改现有容器。 不要修改日志文件的路径，两个容器都必须通过 /var/log/legacy-app.log 来访问该文件。
考点:pod 两个容器共享存储卷 apiVersion: v1
kind: Pod
metadata:
name: podname spec: containers:
- name: count image: busybox args:
- /bin/sh
- -c
  ->
  i=0;
  while true; do
  echo "$(date) INFO $i" >> /var/log/legacy-ap.log; i=$((i+1));
  sleep 1;
  done
  volumeMounts:
- name: logs
  mountPath: /var/log
- name: count-log-1
  image: busybox
  args: [/bin/sh, -c, 'tail -n+1 -f /var/log/legacy-ap.log']
  volumeMounts:
- name: varlog
  mountPath: /var/log
  volumes:
- name: logs
  emptyDir: {}
#验证:
$ kubectl logs <pod_name> -c <container_name>



(10)查询集群中指定 Pod 的 log日志，将带有 Error 的行输出到指定文件
15、设置配置环境:问题权重: 5%
kubectl config use-context k8s
Task
监控 pod bar 的日志并:
提取与错误 file-not-found 相对应的日志行
将这些日志行写入 /opt/KUTR00101/bar
考点:kubectl logs 命令
kubectl logs foobar | grep unable-access-website > /opt/KUTR00101/foobar




(11)1.创建一个 Deployment，2.更新镜像版本，3.回滚？



(12)集群有一个节点 notready，找出问题，并解决。 并保证机器重启后不会再出现此问题？



(13)创建一个 PV，使用hostPath 存储，大小1G，ReadWriteOnce？
12、设置配置环境:问题权重: 4%
kubectl config use-context hk8s Task
创建名为 app-config 的 persistent volume，容量为 1Gi，访问模式为 ReadWriteMany。 volume 类型为 hostPath，位于 /srv/app-config
考点:hostPath 类型的 pv apiVersion: v1
kind: PersistentVolume metadata:
name: app-config labels:
type: local
spec:
capacity:
storage: 2Gi
accessModes:
- ReadWriteMany
  hostPath:
  path: "/src/app-config"




(14)使用指定 storageclass 创建一个 pvc，大小为 10M，将这个 nginx 容器的/var/nginx/html目录使用该 pvc 挂在出来，将这个 pvc 的大小从 10M 更新成 70M?
```yaml
#解答
#创建 PVC
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
name: pv-volume
spec:
accessModes:
- ReadWriteOnce
volumeMode: Filesystem
resources:
requests:
storage: 10Mi
storageClassName: csi-hostpath-sc
#创建 pod
apiVersion: v1
kind: Pod
metadata:
name: web-server
spec:
containers:
- name: nginx
image: nginx
volumeMounts:
- mountPath: "/usr/share/nginx/html"
name: pv-volume
volumes:
- name: pv-volume
persistentVolumeClaim:
claimName: myclaim
#通过 kubectl edit pvc pv-volume 可以进行修改容量
```



14、 问题权重: 7%
名称:web-server
Image:nginx
挂载路径:/usr/share/nginx/html
配置新的 Pod，以对 volume 具有 ReadWriteOnce 权限。
考点:pod 中对 pv 和 pvc 的使用






13、设置配置环境:问题权重: 7%
kubectl config use-context ok8s
Task
创建一个新的 PersistentVolumeClaim:
名称: pv-volume
Class: csi-hostpath-sc
容量: 10Mi
创建一个新的 Pod，此 Pod 将作为 volume 挂载到 PersistentVolumeClaim:
最后，使用 kubectl edit 或 kubectl patch 将 PersistentVolumeClaim 的容量扩 展为 70Mi，并记录此更改。

考点:pvc 的创建 class 属性的使用，--save-config 记录变更 #创建 PVC
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
name: pv-volume
spec:
accessModes:
- ReadWriteOnce
  volumeMode: Filesystem resources:
  requests:
  storage: 10Mi
  storageClassName: csi-hostpath-sc #创建 pod
  apiVersion: v1
  kind: Pod
  metadata:
  name: web-server
  spec:
  containers:
- name: nginx
  image: nginx
  volumeMounts:
- mountPath: "/usr/share/nginx/html" name: pv-volume
  volumes:
- name: pv-volume
  persistentVolumeClaim:
  claimName: myclaim
  
kubectl edit pvc pv-volume --save-config



(15)将集群中一个 Deployment 服务暴露出来,(是一个 nginx，使用kubectl expose 命令暴露即可)？
6、设置配置环境:问题权重: 7%
kubectl config use-context k8s
Task 请重新配置现有的部署 front-end 以及添加名为 http 的端口规范来公开现 有容器 nginx 的端口 80/tcp。
创建一个名为 front-end-svc 的新服务，以公开容器端口 http。 配置此服务，以 通过在排定的节点上的 NodePort 来公开各个 Pods 考点:将现有的 deploy 暴露成 nodeport 的 service。
$ kubectl expose deployment front-end --name=front-end-svc --port=80 -- tarport=80 --type=NodePort



(16)查询集群中节点，找出可以调度节点的数量，(其实就是被标记为不可调度和 打了污点的节点之外的节点 数量 )，然后将数量写到指定文件？
```yaml
#解答
# 查询集群 Ready 节点数量
$ kubectl get node | grep -i ready
# 判断节点有误不可调度污点
$ kubectl describe nodes <nodeName> | grep -i taints | grep -i noSchedule
```


10、设置配置环境:问题权重: 4%
kubectl config use-context k8s
Task
检查有多少 worker nodes 已准备就绪(不包括被打上 Taint:NoSchedule 的节点)， 并将数量写入 /opt/KUSC00402/kusc00402.txt
考点:检查节点角色标签，状态属性，污点属性的使用
$ kubectl describe nodes <nodeName> | grep -i taints | grep -i noSchedule




(17)找集群中带有指定 label 的 Pod 中占用资源最高的，并将它的名字写入指定的文件？
```yaml
#解答
$ kubectl top pod -l name=cpu-user -A
NAMAESPACE NAME CPU MEM
delault cpu-user-1 45m 6Mi
delault cpu-user-2 38m 6Mi
delault cpu-user-3 35m 7Mi
delault cpu-user-4 32m 10Mi

# echo 'cpu-user-1' >>/opt/KUTR00401/KUTR00401.txt
```

17、设置配置环境，问题权重: 5%
kubectl config use-context k8s
Task
通过 pod label name=cpu-loader，找到运行时占用大量 CPU 的 pod， 并将 占用 CPU 最高的 pod 名称写入文件 /opt/KUTR000401/KUTR00401.txt(已 存在)。
考点:kubectl top --l 命令的使用 kubectl top pod -l name=cpu-user -A




(18)创建一个名为 app-config 的 PV，PV 的容量为 2Gi 访问模式为 ReadWriteMany，volume 的类型为 hostPath，位置为/src/app-config？
```yaml
# 解答
apiVersion: v1
kind: PersistentVolume 
metadata:
  name: app-config 
  labels:
    type: local 
spec:
  capacity: 
    storage: 2Gi
  accessModes:
    - ReadWriteMany
  hostPath:
    path: "/src/app-config"
```


(19)将 deployment 扩容到 6 个 pod？
```shell
#解答
# kubectl scale --replicas=6 deployment/loadbalancer
```


(20)创建 NetworkPolicy？
```yaml
apiVersion: networking.k8s.io/v1 
kind: NetworkPolicy
metadata:
  name: all-port-from-namespace
  namespace: internal 
spec:
  podSelector: 
    matchLabels: {}
  ingress: 
    - from:
      - podSelector: {} 
      ports:
        - port: 9000
```

5、设置配置环境:问题权重: 7% kubectl config use-context hk8s
Task
创建一个名为 allow-port-from-namespace 的新 NetworkPolicy，以允许现有 namespace corp-net 中的 Pods 连接到同一 namespace 中其他 Pods 的端 口 9200。
确保新的 NetworkPolicy:
不允许对没有在监听端口 9200 的 Pods 的访问
不允许不来自 namespacecorp-net 的 Pods 的访问
考点:NetworkPolicy 的创建
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
name: all-port-from-namespace
namespace: internal
spec:
podSelector:
matchLabels: {}
ingress:
- from:
- namespaceSelector: matchLabels: name: namespacecorp-net
- podSelector: {}
  ports:
- port: 9000


9、设置配置环境:问题权重: 4%
kubectl config use-context k8s
Task
按如下要求调度一个 pod:
名称:nginx-kusc00401
Image:nginx
Node selector:disk=spinnin
考点:nodeSelect 属性的使用
apiVersion: v1
kind: Pod
metadata:
name: nginx-kusc00401
labels:
role: nginx-kusc00401
spec:
nodeSelector:
disk: spinnin
containers:
- name: nginx
  image: nginx














# CKA 20190714考试真题

# 1.监控 foobar Pod 的日志，提取 pod 相应的行'error'写入到/logs 文件中

```
  kubectl logs foobar | grep error > /logs
```

# 2.使用 name 排序列出所有的 PV，把输出内容存储到/opt/文件中 使用 kubectl own 对输出进行排序，并且不再进一步操作它

```
  kubectl get pv --all-namespace --sort-by=.metadata.name > /opt/
```

# 3.确保在 kubectl 集群的每个节点上运行一个 Nginx Pod。其中 Nginx Pod 必须使用 Nginx 镜像。不要覆盖当前环境中的任何 taints。 使用 Daemonset 来完成这个任务，Daemonset 的名字使用 ds。

	题目对应文档：https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/
	删除tolerations字段，复制到image: gcr.io/fluentd-elasticsearch/fluentd:v2.5.1这里即可，再按题意更改yaml文件。
	apiVersion: apps/v1
	kind: DaemonSet
	metadata:
	  name: ds
	  namespace: kube-system
	  labels:
		k8s-app: fluentd-logging
	spec:
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

# 4.添加一个 initcontainer 到 lum(/etc/data)这个 initcontainer 应该创建一个名为/workdir/calm.txt 的空文件，如果/workdir/calm.txt 没有被检测到，这个 Pod 应该退出

- - ```
      题目中yaml文件已经给出，只需要增加initcontainers部分，以及emptyDir: {} 即可
      init文档位置：https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
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
        command: 
        - touch
        - /workdir/calm.txt
          volumeMounts:
        - name: work
          mountPath: /workdir
        
```



# 5.创建一个名为 kucc 的 Pod,其中内部运行着 nginx+redis+memcached+consul 4 个容器

```
https://v1-14.docs.kubernetes.io/docs/concepts/workloads/pods/pod-overview/
	apiVersion: v1
	kind: Pod
	metadata:
	  name: kucc
	spec:
	  containers:
	  - name: nginx
		image: nginx
	  - name: redis
		image: redis
	  - name: memcached
		image: memcached
	  - name: consul
		image: consul
```

# 6.创建 Pod，名字为 nginx，镜像为 nginx，添加 label disk=ssd

```
https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
apiVersion: v1
	kind: Pod
	metadata:
	  name: nginx
	  labels:
		env: test
	spec:
	  containers:
	  - name: nginx
		image: nginx
		imagePullPolicy: IfNotPresent
	  nodeSelector:
		disk: ssd
```

# 7.创建 deployment 名字为 nginx-app 容器采用 1.11.9 版本的 nginx  这个 deployment 包含 3 个副本,接下来通过滚动升级的方式更新镜像版本为 1.12.0，并记录这个更新，最后，回滚这个更新到之前的 1.11.9 版本

	kubectl run deployment nginx-app --image=nginx:1.11.9 --replicas=3
	kubectl set image deployment nginx-app nginx-app=nginx:1.12.0 --record  (nginx-app container名字)
	kubectl rollout history deployment nginx-app
	kubectl rollout undo deployment nginx-app

# 8.创建和配置 service，名字为 front-end-service。可以通过 NodePort/ClusterIp 开访问，并且路由到 front-end 的 Pod 上

```
kubectl expose pod front-end --name=front-end-service --port=80  --type=NodePort
```

# 9.创建一个 Pod，名字为 Jenkins，镜像使用 Jenkins。在新的 namespace website-frontend 上创建

	kubectl create ns website-frontend
	
	apiVersion: v1
	kind: Pod
	metadata:
	  name: Jenkins
	  namespace: website-frontend
	spec:
	  containers:
	  - name: Jenkins
		image: Jenkins
		
	kubectl apply -f ./xxx.yaml 	

# 10.创建 deployment 的 spec 文件: 使用 redis 镜像，7 个副本，label 为 app_enb_stage=dev deployment 名字为 kual00201 保存这个 spec 文件到/opt/KUAL00201/deploy_spec.yaml完成后，清理(删除)在此任务期间生成的任何新的 k8s API 对象

```
kubectl run kual00201 --image=redis --labels=app_enb_stage=dev --dry-run -oyaml > /opt/KUAL00201/deploy_spec.yaml
```

# 11.创建一个文件/opt/kucc.txt ，这个文件列出所有的 service 为 foo ,在 namespace 为 production 的 Pod这个文件的格式是每行一个 Pod的名字

```
kubectl get svc -n production --show-labels | grep foo

kubectl get pods -l app=foo(label标签) -o=custom-columns=NAME:.spec.name > kucc.txt
```

# 12.创建一个secret,名字为super-secret包含用户名bob,创建pod1挂载该secret，路径为/secret，创建pod2，使用环境变量引用该secret，该变量的环境变量名为ABC

	https://kubernetes.io/zh/docs/concepts/configuration/secret/#%E8%AF%A6%E7%BB%86
	echo -n "bob" | base64
	
	apiVersion: v1
	kind: Secret
	metadata:
	  name: super-secret
	type: Opaque
	data:
	  username: Ym9i
	  
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
	restartPolicy: Never

# 13.在新的ns中创建pv，指定pv名字和挂载路径，镜像等

	https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#create-a-persistentvolume
	kubectl create ns new
	
	apiVersion: v1
	kind: PersistentVolume
	metadata:
	  name: pv0003
	spec:
	  capacity:
	    storage: 5Gi
	volumeMode: Filesystem
	accessModes:
	- ReadWriteOnce
	persistentVolumeReclaimPolicy: Recycle
	storageClassName: slow
	hostPath:
	  path: "/etc/data"
	
	kubectlc apply -f ./xxx.yaml --namespace=new

# 14.为给定deploy  website副本扩容到6

```
 kubectl scale deployment website --replicas=6
```

# 15.查看给定集群ready的node个数(不包含NoSchedule)

```
1.kubectl get nodes 
2.把所有ready得都执行kubectl describe node $nodename | grep Taint  如果有NoSchedule
```

# 16.找出指定ns中使用cup最高的pod名写出到指定文件

```
   kubectc top pod -l xxx --namespace=xxx
```

# 17.创建一个 deployment 名字为:nginx-dns 路由服务名为：nginx-dns 确保服务和 pod 可以通过各自的 DNS 记录访问 容器使用 nginx 镜像，使用 nslookup 工具来解析 service 和 pod 的记录并写入相应的/opt/service.dns 和/opt/pod.dns 文件中，确保你使用 busybox:1.28 的镜像用来测试。

- ```
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

# 18.给定https地址，ca，cert证书，key备份该数据到指定目录

```
ETCDCTL_API=3 etcdctl --endpoints=https://127.0.0.1:1111 --ca-file=/pki/ca.crt --cert-file=/pki/cert.crt --key-file=/pki/key.crt snapshot save 给的路径
有些题目下--ca-file会报错，记得看endpoints -h 里的字段怎么要求的
```

# 19.在ek8s集群中使name=ek8s-node-1节点不能被调度，并使已被调度的pod重新调度

```
先切换集群到ek8s    
再执行
kubectl drain node1 --ignore-daemonsets --delete-local-data  
```

# 20.给定集群中的一个node未处于ready状态，解决该问题并具有持久性

```
进入集群
ssh node  

systemctl status kubelet

systemctl start kubelet   
systemctl enable kubelet
```

# 21.题目很绕，大致是 在k8s的集群中的node1节点配置kubelet的service服务，去拉起一个由kubelet直接管理的pod(说明了是静态pod)，

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

# 22.某集群中kubelet.service服务无法正常启动，解决该问题，并具有持久性

```
kubectl 命令能用 kubectl get cs 健康检查  看manager-controller  是否ready   如果不ready   systemctl start kube-manager-controller.service   
```


23.TLS问题 （一道很长的题目，建议放弃，难度特别大）

# 24.创建指定大小和路径的pv

```
https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#create-a-persistentvolume
```



