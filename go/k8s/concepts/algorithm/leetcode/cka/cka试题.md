

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


