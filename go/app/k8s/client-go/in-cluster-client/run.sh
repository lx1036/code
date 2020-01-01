# build
GOOS=linux go build -o ./bin/app ./main.go

# If you are running a Minikube cluster, you can build this image directly on the Docker engine of the Minikube node without pushing it to a registry. To build the image on Minikube.
eval "$(minikube docker-env)"
docker build -t in-cluster .

# If you have RBAC enabled on your cluster, use the following snippet to create role binding which will grant the default service account view permissions.
kubectl create clusterrolebinding default-view --clusterrole=view --serviceaccount=default:default

# run the image in a Pod with a single instance Deployment
kubectl run --rm -i demo --image=in-cluster --image-pull-policy=Never
