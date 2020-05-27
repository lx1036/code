

import {Component} from '@angular/core';
import {
  CronJobList,
  DaemonSetList,
  DeploymentList,
  JobList,
  Metric,
  PodList,
  ReplicaSetList,
  ReplicationControllerList,
  StatefulSetList,
} from '@api/backendapi';
import {OnListChangeEvent, ResourcesRatio} from '@api/frontendapi';

import {ListGroupIdentifier, ListIdentifier} from '../common/components/resourcelist/groupids';
import {emptyResourcesRatio} from '../common/components/workloadstatus/component';
import {GroupedResourceList} from '../common/resources/groupedlist';

import {Helper, ResourceRatioModes} from './helper';

@Component({
  selector: 'kd-overview',
  templateUrl: './template.html',
})
export class OverviewComponent extends GroupedResourceList {
  hasWorkloads(): boolean {
    return this.isGroupVisible(ListGroupIdentifier.workloads);
  }

  hasDiscovery(): boolean {
    return this.isGroupVisible(ListGroupIdentifier.discovery);
  }

  hasConfig(): boolean {
    return this.isGroupVisible(ListGroupIdentifier.config);
  }

  showWorkloadStatuses(): boolean {
    return (
      Object.values(this.resourcesRatio).reduce((sum, ratioItems) => sum + ratioItems.length, 0) !==
      0
    );
  }
}
