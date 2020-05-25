

**[Calico the hard way](https://docs.projectcalico.org/getting-started/kubernetes/hardway/overview)**:
1. install kubernetes
2. install calico datastore: Calico stores the data about the operational and configuration state of your cluster in a central datastore.
3. install ip pools
4. install cni plugin: Kubernetes uses the Container Network Interface (CNI) to interact with networking providers like Calico. 且 cni plugin
必须以 DeamonSet 形式安装在每一个 Node 节点上。The CNI plugin interacts with the Kubernetes API server while creating pods, 
both to obtain additional information and to update the datastore with information about the pod.


