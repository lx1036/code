

import {Component, Input} from '@angular/core';
import {ResourcesRatio} from '@api/frontendapi';

export const emptyResourcesRatio: ResourcesRatio = {
  cronJobRatio: [],
  daemonSetRatio: [],
  deploymentRatio: [],
  jobRatio: [],
  podRatio: [],
  replicaSetRatio: [],
  replicationControllerRatio: [],
  statefulSetRatio: [],
};

@Component({
  selector: 'kd-workload-statuses',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class WorkloadStatusComponent {
  @Input() resourcesRatio = emptyResourcesRatio;
  colors: string[] = [];

  getCustomColor(label: string): string {
    if (label.includes('Running')) {
      return '#00c752';
    } else if (label.includes('Succeeded')) {
      return '#006028';
    } else if (label.includes('Pending')) {
      return '#ffad20';
    } else if (label.includes('Failed')) {
      return '#f00';
    } else {
      return '';
    }
  }
}
