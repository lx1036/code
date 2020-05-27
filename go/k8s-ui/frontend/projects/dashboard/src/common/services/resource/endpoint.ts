

const baseHref = 'api/v1';

export enum Resource {
  job = 'job',
  cronJob = 'cronjob',
  crd = 'crd',
  crdFull = 'customresourcedefinition',
  crdObject = 'object',
  daemonSet = 'daemonset',
  deployment = 'deployment',
  pod = 'pod',
  replicaSet = 'replicaset',
  oldReplicaSet = 'oldreplicaset',
  newReplicaSet = 'newreplicaset',
  horizontalPodAutoscaler = 'horizontalpodautoscaler',
  replicationController = 'replicationcontroller',
  statefulSet = 'statefulset',
  node = 'node',
  namespace = 'namespace',
  persistentVolume = 'persistentvolume',
  storageClass = 'storageclass',
  clusterRole = 'clusterrole',
  configMap = 'configmap',
  persistentVolumeClaim = 'persistentvolumeclaim',
  secret = 'secret',
  ingress = 'ingress',
  service = 'service',
  event = 'event',
  container = 'container',
  plugin = 'plugin',
}

export enum Utility {
  shell = 'shell',
}

class ResourceEndpoint {
  constructor(private readonly resource_: Resource, private readonly namespaced_ = false) {}

  list(): string {
    return `${baseHref}/${this.resource_}${this.namespaced_ ? '/:namespace' : ''}`;
  }

  detail(): string {
    return `${baseHref}/${this.resource_}${this.namespaced_ ? '/:namespace' : ''}/:name`;
  }

  child(resourceName: string, relatedResource: Resource, resourceNamespace?: string): string {
    if (!resourceNamespace) {
      resourceNamespace = ':namespace';
    }

    return `${baseHref}/${this.resource_}${
      this.namespaced_ ? `/${resourceNamespace}` : ''
    }/${resourceName}/${relatedResource}`;
  }
}

class UtilityEndpoint {
  constructor(private readonly utility_: Utility) {}

  shell(namespace: string, resourceName: string): string {
    return `${baseHref}/${Resource.pod}/${namespace}/${resourceName}/${this.utility_}`;
  }
}

export class EndpointManager {
  static resource(resource: Resource, namespaced?: boolean): ResourceEndpoint {
    return new ResourceEndpoint(resource, namespaced);
  }

  static utility(utility: Utility): UtilityEndpoint {
    return new UtilityEndpoint(utility);
  }
}
