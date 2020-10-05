import {TemplateState} from "../shared.const";
import {
  KubePersistentVolumeClaim,
  PersistentVolumeClaimFilesystemStatus,
  PersistentVolumeClaimSnap
} from "./kubernetes/persistentvolumeclaim";


export class PublishStatus {
  id: number;
  type: number;
  resourceId: number;
  templateId: number;
  cluster: string;
  state: TemplateState;
  pvc: KubePersistentVolumeClaim;
  fileSystemStatus: PersistentVolumeClaimFilesystemStatus;
  rbdImage: string;
  snaps: PersistentVolumeClaimSnap[];
  errNum: number;
}


