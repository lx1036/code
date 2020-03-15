

kubectl get pods -l run=busybox
POD_NAME=$(kubectl get pods -l run=busybox -o jsonpath="{.items[0].metadata.name}")
kubectl exec -ti "$POD_NAME" -- nslookup kubernetes



#NAME                       READY   STATUS    RESTARTS   AGE
#busybox-6f584dc999-qfk8b   1/1     Running   0          5m22s
#Server:         10.32.0.10
#Address:        10.32.0.10:53
#
#Name:   kubernetes.default.svc.cluster.local
#Address: 10.32.0.1
#
#*** Can't find kubernetes.svc.cluster.local: No answer
#*** Can't find kubernetes.cluster.local: No answer
#*** Can't find kubernetes.asia-east1-a.c.stable-framing-241507.internal: No answer
#*** Can't find kubernetes.c.stable-framing-241507.internal: No answer
#*** Can't find kubernetes.google.internal: No answer
#*** Can't find kubernetes.default.svc.cluster.local: No answer
#*** Can't find kubernetes.svc.cluster.local: No answer
#*** Can't find kubernetes.cluster.local: No answer
#*** Can't find kubernetes.asia-east1-a.c.stable-framing-241507.internal: No answer
#*** Can't find kubernetes.c.stable-framing-241507.internal: No answer
#*** Can't find kubernetes.google.internal: No answer
