kubectl run nginx --image=nginx
kubectl get pods -l run=nginx

POD_NAME=$(kubectl get pods -l run=nginx -o jsonpath="{.items[0].metadata.name}")
kubectl port-forward "${POD_NAME}" 8080:80

curl --head http://127.0.0.1:8080

kubectl logs "${POD_NAME}"
kubectl exec -ti "${POD_NAME}" -- nginx -v

kubectl expose deployment nginx --port 80 --type NodePort
NODE_PORT=$(kubectl get svc nginx \
  --output=jsonpath='{range .spec.ports[0]}{.nodePort}')
EXTERNAL_IP=$(gcloud compute instances describe worker-0 \
  --format 'value(networkInterfaces[0].accessConfigs[0].natIP)')

curl -I http://"${EXTERNAL_IP}":"${NODE_PORT}"

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
POD_ID=$(sudo crictl -r unix:///var/run/containerd/containerd.sock \
  pods --name untrusted -q)

CONTAINER_ID=$(sudo crictl -r unix:///var/run/containerd/containerd.sock \
  ps -p "${POD_ID}" -q)

sudo runsc --root /run/containerd/runsc/k8s.io ps "${CONTAINER_ID}"
