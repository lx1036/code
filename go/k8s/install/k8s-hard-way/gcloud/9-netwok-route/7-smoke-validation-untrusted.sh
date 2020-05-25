

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: untrusted
  annotations:
    io.kubernetes.cri.untrusted-workload: "true"
spec:
  containers:
    - name: webserver
      image: gcr.io/hightowerlabs/helloworld:2.0.0
EOF

kubectl get pods -o wide

sleep 10s

INSTANCE_NAME=$(kubectl get pod untrusted --output=jsonpath='{.spec.nodeName}')
gcloud compute ssh "${INSTANCE_NAME}"
sudo runsc --root  /run/containerd/runsc/k8s.io list
POD_ID=$(sudo crictl -r unix:///var/run/containerd/containerd.sock \
  pods --name untrusted -q)
CONTAINER_ID=$(sudo crictl -r unix:///var/run/containerd/containerd.sock \
  ps -p "${POD_ID}" -q)
sudo runsc --root /run/containerd/runsc/k8s.io ps "${CONTAINER_ID}"
