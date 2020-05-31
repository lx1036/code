

import {Component} from '@angular/core';
import {ListGroupIdentifier} from '../common/components/resourcelist/groupids';
import {GroupedResourceList} from '../common/resources/groupedlist';

@Component({selector: 'kd-search', templateUrl: './template.html'})
export class SearchComponent extends GroupedResourceList {
  hasCluster(): boolean {
    return this.isGroupVisible(ListGroupIdentifier.cluster);
  }

  hasWorkloads(): boolean {
    return this.isGroupVisible(ListGroupIdentifier.workloads);
  }

  hasDiscovery(): boolean {
    return this.isGroupVisible(ListGroupIdentifier.discovery);
  }

  hasConfig(): boolean {
    return this.isGroupVisible(ListGroupIdentifier.config);
  }
}
