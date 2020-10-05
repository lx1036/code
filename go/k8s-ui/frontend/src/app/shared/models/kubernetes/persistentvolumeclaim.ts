
export class StatusCause {
  reason: string;
  message: string;
  field: string;
}

export class StatusDetails {
  name: string;
  group: string;
  kind: string;
  uid: string;
  causes: StatusCause[];
  retryAfterSeconds: number;
}

export class ListMeta {
  selfLink: string;
  resourceVersion: string;
}

export class Status {
  kind: string;
  apiVersion: string;
  metadata: ListMeta;
  status: string;
  message: string;
  reason: string;
  details: StatusDetails;
  code: number;
}

export class Initializer {
  name: string;
}

export class Initializers {
  pending: Initializer[];
  result: Status;
}

export class ObjectMeta {
  name: string;
  generateName: string;
  namespace: string;
  selfLink: string;
  uid: string;
  resourceVersion: string;
  generation: number;
  creationTimestamp: Time;
  deletionTimestamp: Time;
  deletionGracePeriodSeconds: number;
  labels: {};
  annotations: {};
  ownerReferences: OwnerReference[];
  initializers: Initializers;
  finalizers: string[];
  clusterName: string;
}

export class OwnerReference {
  apiVersion: string;
  kind: string;
  name: string;
  uid: string;
  controller: boolean;
  blockOwnerDeletion: boolean;
}

export class Time {
}

export class ResourceRequirements {
  limits: {};
  requests: {};
}

export class LabelSelectorRequirement {
  key: string;
  operator: string;
  values: string[];
}

export class LabelSelector {
  matchLabels: {};
  matchExpressions: LabelSelectorRequirement[];
}

export class PersistentVolumeClaimSpec {
  accessModes: string[];
  selector: LabelSelector;
  resources: ResourceRequirements;
  volumeName: string;
  storageClassName: string;
}

export class PersistentVolumeClaimStatus {
  phase: string;
  accessModes: string[];
  capacity: {};
}

export class KubePersistentVolumeClaim {
  kind: string;
  apiVersion: string;
  metadata: ObjectMeta;
  spec: PersistentVolumeClaimSpec;
  status: PersistentVolumeClaimStatus;
}

export class PersistentVolumeClaimFilesystemStatus {
  msg: string;
  status: string[];
  rbdImage: string;
  imageType: string;
}

export class PersistentVolumeClaimSnap {
  id: number;
  name: string;
  size: number;
}

