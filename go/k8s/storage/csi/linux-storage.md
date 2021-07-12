
## Block Device
Block Device 和 File System 区别联系：
File System 如 ext4 type，是在 Block Device 基础上构建的，块存储 Block Device 就是一块SSD硬盘，固定size的随机访问。
一般用于数据库这种软件。

### CSI 如何支持 Block Device 块存储
https://kubernetes-csi.github.io/docs/raw-block.html

如何在 pod 内使用 rbd: https://kubernetes.io/blog/2019/03/07/raw-block-volume-support-to-beta/#using-a-raw-block-pvc
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
spec:
  accessModes:
    - ReadWriteMany
  volumeMode: Block
  storageClassName: my-sc
  resources:
    requests:
      storage: 1Gi

---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
    - name: my-container
      image: busybox
      command:
        - sleep
        - "3600"
      volumeDevices: # 这里是重点
        - devicePath: /dev/block
          name: my-volume
      imagePullPolicy: IfNotPresent
  volumes:
    - name: my-volume
      persistentVolumeClaim:
        claimName: my-pvc
```
