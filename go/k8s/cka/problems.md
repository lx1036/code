**[kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)**
**[真题1](https://blog.csdn.net/deerjoe/java/article/details/86300826)**

# Storage 7%

## (1)列出集群所有的pv，并以 name 字段排序（使用kubectl自带排序功能）
```shell script
kubectl get pv --sort-by=.metadata.name
```
### (1.1)创建一个1G可读可写的PV，挂载在宿主机的"/mnt/data"目录
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: task-pv-volume
  labels:
    type: local
spec:
  storageClassName: manual
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/mnt/data"
```
 

列出指定pod的日志中状态为Error的行，并记录在指定的文件上

kubectl logs <podname> | grep bash > /opt/KUCC000xxx/KUCC000xxx.txt
1
考点：Monitor, Log, and Debug

列出k8s可用的节点，不包含不可调度的 和 NoReachable的节点，并把数字写入到文件里

#笨方法，人工数
kubectl get nodes

#CheatSheet方法，应该还能优化JSONPATH
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}' \
 && kubectl get nodes -o jsonpath="$JSONPATH" | grep "Ready=True"
1
2
3
4
5
6
考点：kubectl命令熟悉程度

参考：kubectl cheatsheet

创建一个pod名称为nginx，并将其调度到节点为 disk=stat上

#我的操作,实际上从文档复制更快
kubectl run nginx --image=nginx --restart=Never --dry-run > 4.yaml
#增加对应参数
vi 4.yaml
kubectl apply -f 4.yaml
1
2
3
4
5
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
    disktype: ssd
1
2
3
4
5
6
7
8
9
10
11
12
13
考点：pod的调度。

参考：assign-pod-node

提供一个pod的yaml，要求添加Init Container，Init Container的作用是创建一个空文件，pod的Containers判断文件是否存在，不存在则退出

apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: apline
    image: nginx
    command: ['sh', '-c', 'if [目录下有work文件];then sleep 3600; else exit; fi;']
###增加init Container####
initContainers:
 - name: init
    image: busybox
    command: ['sh', '-c', 'touch /目录/work;']
    
1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
考点：init Container。一开始审题不仔细，以为要用到livenessProbes

参考：init-containers

指定在命名空间内创建一个pod名称为test，内含四个指定的镜像nginx、redis、memcached、busybox

kubectl run test --image=nginx --image=redis --image=memcached --image=buxybox --restart=Never -n <namespace>
1
考点：kubectl命令熟悉程度、多个容器的pod的创建

参考：kubectl cheatsheet

创建一个pod名称为test，镜像为nginx，Volume名称cache-volume为挂在在/data目录下，且Volume是non-Persistent的

apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  containers:
  - image: nginx
    name: test-container
    volumeMounts:
    - mountPath: /cache
      name: cache-volume
  volumes:
  - name: cache-volume
    emptyDir: {}
1
2
3
4
5
6
7
8
9
10
11
12
13
14
考点：Volume、emptdir

参考：Volumes

列出Service名为test下的pod 并找出使用CPU使用率最高的一个，将pod名称写入文件中

#使用-o wide 获取service test的SELECTOR
kubectl get svc test -o wide
##获取结果我就随便造了
NAME              TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE       SELECTOR
test   ClusterIP   None         <none>        3306/TCP   50d       app=wordpress,tier=mysql

#获取对应SELECTOR的pod使用率，找到最大那个写入文件中
kubectl top test -l 'app=wordpress,tier=mysql'
1
2
3
4
5
6
7
8
考点：获取service selector，kubectl top监控pod资源

参考：Tools for Monitoring Resources

创建一个Pod名称为nginx-app，镜像为nginx，并根据pod创建名为nginx-app的Service，type为NodePort

kubectl run nginx-app --image=nginx --restart=Never --port=80
kubectl create svc nodeport nginx-app --tcp=80:80 --dry-run -o yaml > 9.yaml
#修改yaml，保证selector name=nginx-app
vi 9.yaml
1
2
3
4
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: null
  labels:
    app: nginx-app
  name: nginx-app
spec:
  ports:
  - name: 80-80
    port: 80
    protocol: TCP
    targetPort: 80
  selector:
#注意要和pod对应  
    name: nginx-app
  type: NodePort
1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
考点：Service

参考：publishing-services-service-types

创建一个nginx的Workload，保证其在每个节点上运行，注意不要覆盖节点原有的Tolerations

这道题直接复制文档的yaml太长了，由于damonSet的格式和Deployment格式差不多，我用旁门左道的方法 先创建Deploy，再修改，这样速度会快一点

#先创建一个deployment的yaml模板
kubectl run nginx --image=nginx --dry-run -o yaml > 10.yaml
#将yaml改成DaemonSet
vi 10.yaml
1
2
3
4
#修改apiVersion和kind
#apiVersion: extensions/v1beta1
#kind: Deployment
apiVersion:apps/v1
kind: DaemonSet
metadata:
  creationTimestamp: null
  labels:
    run: nginx
  name: nginx
spec:
#去掉replicas
# replicas: 1
  selector:
    matchLabels:
      run: nginx
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        run: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
        resources: {}
status: {}

1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
18
19
20
21
22
23
24
25
26
27
28
29
考点：DaemonSet

参考：DaemonSet

将deployment为nginx-app的副本数从1变成4。

#方法1
kubectl scale  --replicas=4 deployment nginx-app
#方法2，使用edit命令将replicas改成4
kubectl edit deploy nginx-app
1
2
3
4
考点：deployment的Scaling，搜索Scaling

参考：Scaling the application by increasing the replica count

创建nginx-app的deployment ，使用镜像为nginx:1.11.0-alpine ,修改镜像为1.11.3-alpine，并记录升级，再使用回滚，将镜像回滚至nginx:1.11.0-alpine

kubectl run nginx-app --image=nginx:1.11.0-alpine
kubectl set image deployment nginx-app --image=nginx:1.11.3-alpine
kubectl rollout undo deployment nginx-app 
kubectl rollout status -w deployment nginx-app 
1
2
3
4
考点：资源的更新

参考：Kubectl Cheat Sheet:Updating Resources

根据已有的一个nginx的pod、创建名为nginx的svc、并使用nslookup查找出service dns记录，pod的dns记录并分别写入到指定的文件中

#创建一个服务
kubectl create svc nodeport nginx --tcp=80:80
#创建一个指定版本的busybox，用于执行nslookup
kubectl create -f https://k8s.io/examples/admin/dns/busybox.yaml
#将svc的dns记录写入文件中
kubectl exec -ti busybox -- nslookup nginx > 指定文件
#获取pod的ip地址
kubectl get pod nginx -o yaml
#将获取的pod ip地址使用nslookup查找dns记录
kubectl exec -ti busybox -- nslookup <Pod ip>
1
2
3
4
5
6
7
8
9
10
考点：网络相关，DNS解析

参考：Debugging DNS Resolution

创建Secret 名为mysecret，内含有password字段，值为bob，然后 在pod1里 使用ENV进行调用，Pod2里使用Volume挂载在/data 下

#将密码值使用base64加密,记录在Notepad里
echo -n 'bob' | base64
1
2
14.secret.yaml

apiVersion: v1
kind: Secret
metadata:
  name: mysecret
type: Opaque
data:
  password: Ym9i
1
2
3
4
5
6
7
14.pod1.yaml

apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  containers:
  - name: mypod
    image: nginx
    volumeMounts:
    - name: mysecret
      mountPath: "/data"
      readOnly: true
  volumes:
  - name: mysecret
    secret:
      secretName: mysecret
1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
14.pod2.yaml

apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  containers:
  - name: mycontainer
    image: redis
    env:
      - name: SECRET_PASSWORD
        valueFrom:
          secretKeyRef:
            name: mysecret
            key: password
1
2
3
4
5
6
7
8
9
10
11
12
13
14
考点 Secret

参考：Secret

使node1节点不可调度，并重新分配该节点上的pod

#直接drain会出错，需要添加--ignore-daemonsets --delete-local-data参数
kubectl drain node node1  --ignore-daemonsets --delete-local-data
1
2
考点：节点调度、维护

参考：[Safely Drain a Node while Respecting Application SLOs]: （https://kubernetes.io/docs/tasks/administer-cluster/safely-drain-node/)

使用etcd 备份功能备份etcd（提供enpoints，ca、cert、key）

ETCDCTL_API=3 etcdctl --endpoints https://127.0.0.1:2379 \
--cacert=ca.pem --cert=cert.pem --key=key.pem \
snapshot save snapshotdb
1
2
3
考点：etcd的集群的备份与恢复

参考：backing up an etcd cluster

给出一个失联节点的集群，排查节点故障，要保证改动是永久的。

#查看集群状态
kubectl get nodes
#查看故障节点信息
kubectl describe node node1

#Message显示kubelet无法访问（记不清了）
#进入故障节点
ssh node1

#查看节点中的kubelet进程
ps -aux | grep kubelete
#没找到kubelet进程，查看kubelet服务状态
systemctl status kubelet.service 
#kubelet服务没启动，启动服务并观察
systemctl start kubelet.service 
#启动正常，enable服务
systemctl enable kubelet.service 

#回到考试节点并查看状态
exit

kubectl get nodes #正常

1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
18
19
20
21
22
23
考点：故障排查

参考：Troubleshoot Clusters

给出一个集群，排查出集群的故障

这道题没空做完。kubectl get node显示connection refuse，估计是apiserver的故障。

考点：故障排查

参考：Troubleshoot Clusters

给出一个节点，完善kubelet配置文件，要求使用systemd配置kubelet

这道题没空做完，

考点我知道，doc没找到··在哪 逃~ε=ε=ε=┏(゜ロ゜;)┛

给出一个集群，将节点node1添加到集群中，并使用TLS bootstrapping

这道题没空做完，花费时间比较长，可惜了。

考点：TLS Bootstrapping

参考：TLS Bootstrapping

创建一个pv，类型是hostPath，位于/data中，大小1G，模式ReadOnlyMany

apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv
spec:
  capacity:
    storage: 1Gi  
  accessModes:
    - ReadOnlyMany
  hostPath:
    path: /data 

1
2
3
4
5
6
7
8
9
10
11
12
考点：创建PV
参考：persistent volumes

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


