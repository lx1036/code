

export enum ListIdentifier {
  clusterRole = 'clusterRoleList',
  namespace = 'namespaceList',
  node = 'nodeList',
  persistentVolume = 'persistentVolumeList',
  storageClass = 'storageClassList',
  cronJob = 'cronJobList',
  crd = 'crdList',
  crdObject = 'crdObjectList',
  job = 'jobList',
  deployment = 'deploymentList',
  daemonSet = 'daemonSetList',
  pod = 'podList',
  horizontalpodautoscaler = 'horizontalPodAutoscalerList',
  replicaSet = 'replicaSetList',
  ingress = 'ingressList',
  service = 'serviceList',
  configMap = 'configMapList',
  persistentVolumeClaim = 'persistentVolumeClaimList',
  secret = 'secretList',
  replicationController = 'replicationControllerList',
  statefulSet = 'statefulSetList',
  event = 'event',
  resource = 'resource',
  plugin = 'plugin',
}

export enum ListGroupIdentifier {
  cluster = 'clusterGroup',
  workloads = 'workloadsGroup',
  discovery = 'discoveryGroup',
  config = 'configGroup',
  none = 'none',
}
