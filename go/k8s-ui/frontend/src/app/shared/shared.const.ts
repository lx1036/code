

export const defaultRoutingUrl = 'portal/namespace/0/app';

export const LoginTokenKey = 'jwt_token';

export const enum AlertType {
  DANGER, WARNING, INFO, SUCCESS
}

export const enum TemplateState {
  SUCCESS, FAILD, NOT_FOUND
}

export const httpStatusCode = {
  NoContent: 204,
  Unauthorized: 401,
  Forbidden: 403,
  NotFound: 404,
  StatusInternalServerError: 500
};

export const AuthType = {
  DB: 'db',
  OAuth2: 'oauth2',
  Ldap: 'ldap',
};

export const enum ConfirmationState {
  NA, CONFIRMED, CANCEL
}

export const enum ConfirmationTargets {
  EMPTY,
  USER,
  GROUP,
  PERMISSION,
  CLUSTER,
  TRASH_CLUSTER,
  NAMESPACE_USER,
  NAMESPACE,
  TRASH_NAMESPACE,
  APP_USER,
  APP,
  TRASH_APP,
  DEPLOYMENT,
  POD,
  TRASH_DEPLOYMENT,
  DEPLOYMENT_TPL,
  TRASH_DEPLOYMENT_TPL,
  SERVICE,
  TRASH_SERVICE,
  SERVICE_TPL,
  TRASH_SERVICE_TPL,
  SERVICE_EDGE_NODE,
  SERVICE_AVAILABLE_PORT,
  SERVICE_USED_PORT,
  INGRESS,
  INGRESS_TPL,
  TRASH_INGRESS,
  TRASH_INGRESS_TPL,
  AUTOSCALE,
  AUTOSCALE_TPL,
  TRASH_AUTOSCALE,
  TRASH_AUTOSCALE_TPL,
  CONFIGMAP,
  TRASH_CONFIGMAP,
  CONFIGMAP_TPL,
  TRASH_CONFIGMAP_TPL,
  SECRET,
  TRASH_SECRET,
  SECRET_TPL,
  TRASH_SECRET_TPL,
  PERSISTENT_VOLUME,
  PERSISTENT_VOLUME_RBD_IMAGES,
  PERSISTENT_VOLUME_CLAIM,
  PERSISTENT_VOLUME_CLAIM_SNAPSHOT,
  PERSISTENT_VOLUME_CLAIM_SNAPSHOT_ROLLBACK,
  PERSISTENT_VOLUME_CLAIM_SNAPSHOT_ALL,
  TRASH_PERSISTENT_VOLUME_CLAIM,
  PERSISTENT_VOLUME_CLAIM_TPL,
  TRASH_PERSISTENT_VOLUME_CLAIM_TPL,
  CRONJOB,
  CRONJOB_TPL,
  TRASH_CRONJOB,
  TRASH_CRONJOB_TPL,
  JOB,
  API_KEY,
  WEBHOOK,
  STATEFULSET,
  TRASH_STATEFULSET,
  STATEFULSET_TPL,
  TRASH_STATEFULSET_TPL,
  CONFIG,
  DAEMONSET,
  TRASH_DAEMONSET,
  DAEMONSET_TPL,
  TRASH_DAEMONSET_TPL,
  NOTIFICATION,
  NODE,
  HPA,
  ENDPOINT
}


export const KubeApiTypeDeployment = 'Deployment';
export const KubeApiTypeCronJob = 'CronJob';
export const KubeApiTypeStatefulSet = 'StatefulSet';
export const KubeApiTypeDaemonSet = 'DaemonSet';
export const KubeApiTypeService = 'Service';
export const KubeApiTypeIngress = 'Ingress';
export const KubeApiTypeConfigMap = 'ConfigMap';
export const KubeApiTypeSecret = 'Secret';
export const KubeApiTypePersistentVolumeClaim = 'PersistentVolumeClaim';
export const KubeApiTypeAutoscale = 'Autoscale';


export const SideNavCollapseStorage = 'nav-collapse';
// 同步发布状态时间
export const syncStatusInterval = 5 * 1000;

export type KubeResourcesName = string;
export const KubeResourceConfigMap: KubeResourcesName = 'configmaps';
export const KubeResourceDaemonSet: KubeResourcesName = 'daemonsets';
export const KubeResourceDeployment: KubeResourcesName = 'deployments';
export const KubeResourceEvent: KubeResourcesName = 'events';
export const KubeResourceHorizontalPodAutoscaler: KubeResourcesName = 'horizontalpodautoscalers';
export const KubeResourceIngress: KubeResourcesName = 'ingresses';
export const KubeResourceJob: KubeResourcesName = 'jobs';
export const KubeResourceCronJob: KubeResourcesName = 'cronjobs';
export const KubeResourceNamespace: KubeResourcesName = 'namespaces';
export const KubeResourceNode: KubeResourcesName = 'nodes';
export const KubeResourcePersistentVolumeClaim: KubeResourcesName = 'persistentvolumeclaims';
export const KubeResourcePersistentVolume: KubeResourcesName = 'persistentvolumes';
export const KubeResourcePod: KubeResourcesName = 'pods';
export const KubeResourceReplicaSet: KubeResourcesName = 'replicasets';
export const KubeResourceSecret: KubeResourcesName = 'secrets';
export const KubeResourceService: KubeResourcesName = 'services';
export const KubeResourceStatefulSet: KubeResourcesName = 'statefulsets';
export const KubeResourceEndpoint: KubeResourcesName = 'endpoints';
export const KubeResourceStorageClass: KubeResourcesName = 'storageclasses';
export const KubeResourceRole: KubeResourcesName = 'roles';
export const KubeResourceRoleBinding: KubeResourcesName = 'rolebindings';
export const KubeResourceClusterRole: KubeResourcesName = 'clusterroles';
export const KubeResourceClusterRoleBinding: KubeResourcesName = 'clusterrolebindings';
export const KubeResourceServiceAccount: KubeResourcesName = 'serviceaccounts';
export const KubeResourceCustomResourceDefinition: KubeResourcesName = 'customresourcedefinitions';
