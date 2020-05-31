

import {HttpParams} from '@angular/common/http';
import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  ComponentFactoryResolver,
  Input,
} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {Event, Metric, ReplicaSet, ReplicaSetList} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';

import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-replica-set-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ReplicaSetListComponent extends ResourceListWithStatuses<ReplicaSetList, ReplicaSet> {
  @Input() title: string;
  @Input() endpoint = EndpointManager.resource(Resource.replicaSet, true).list();
  @Input() showMetrics = false;
  cumulativeMetrics: Metric[];

  constructor(
    private readonly replicaSet_: NamespacedResourceService<ReplicaSetList>,
    private readonly activatedRoute_: ActivatedRoute,
    notifications: NotificationsService,
    resolver: ComponentFactoryResolver,
    cdr: ChangeDetectorRef,
  ) {
    super('replicaset', notifications, cdr, resolver);
    this.id = ListIdentifier.replicaSet;
    this.groupId = ListGroupIdentifier.workloads;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.timelapse, 'kd-muted', this.isInPendingState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  getResourceObservable(params?: HttpParams): Observable<ReplicaSetList> {
    return this.replicaSet_.get(this.endpoint, undefined, undefined, params);
  }

  map(rsList: ReplicaSetList): ReplicaSet[] {
    this.cumulativeMetrics = rsList.cumulativeMetrics;
    return rsList.replicaSets;
  }

  isInErrorState(resource: ReplicaSet): boolean {
    return resource.podInfo.warnings.length > 0;
  }

  isInPendingState(resource: ReplicaSet): boolean {
    return (
      resource.podInfo.warnings.length === 0 &&
      (resource.podInfo.pending > 0 || resource.podInfo.running !== resource.podInfo.desired)
    );
  }

  isInSuccessState(resource: ReplicaSet): boolean {
    return (
      resource.podInfo.warnings.length === 0 &&
      resource.podInfo.pending === 0 &&
      resource.podInfo.running === resource.podInfo.desired
    );
  }

  protected getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'pods', 'created', 'images'];
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }

  hasErrors(replicaSet: ReplicaSet): boolean {
    return replicaSet.podInfo.warnings.length > 0;
  }

  getEvents(replicaSet: ReplicaSet): Event[] {
    return replicaSet.podInfo.warnings;
  }
}
