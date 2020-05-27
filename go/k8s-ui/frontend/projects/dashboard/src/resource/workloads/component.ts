

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
import {emptyResourcesRatio} from 'common/components/workloadstatus/component';
import {Helper, ResourceRatioModes} from 'overview/helper';

import {ListGroupIdentifier, ListIdentifier} from '../../common/components/resourcelist/groupids';
import {GroupedResourceList} from '../../common/resources/groupedlist';

@Component({
  selector: 'kd-workloads',
  templateUrl: './template.html',
})
export class WorkloadsComponent extends GroupedResourceList {
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
