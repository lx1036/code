

import {HttpParams} from '@angular/common/http';
import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  ComponentFactoryResolver,
  Input,
} from '@angular/core';
import {DaemonSet, DaemonSetList, Event, Metric} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';

import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-daemon-set-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DaemonSetListComponent extends ResourceListWithStatuses<DaemonSetList, DaemonSet> {
  @Input() endpoint = EndpointManager.resource(Resource.daemonSet, true).list();
  @Input() showMetrics = false;
  cumulativeMetrics: Metric[];

  constructor(
    private readonly daemonSet_: NamespacedResourceService<DaemonSetList>,
    resolver: ComponentFactoryResolver,
    notifications: NotificationsService,
    cdr: ChangeDetectorRef,
  ) {
    super('daemonset', notifications, cdr, resolver);
    this.id = ListIdentifier.daemonSet;
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

  getResourceObservable(params?: HttpParams): Observable<DaemonSetList> {
    return this.daemonSet_.get(this.endpoint, undefined, undefined, params);
  }

  map(daemonSetList: DaemonSetList): DaemonSet[] {
    this.cumulativeMetrics = daemonSetList.cumulativeMetrics;
    return daemonSetList.daemonSets;
  }

  isInErrorState(resource: DaemonSet): boolean {
    return resource.podInfo.warnings.length > 0;
  }

  isInPendingState(resource: DaemonSet): boolean {
    return resource.podInfo.warnings.length === 0 && resource.podInfo.pending > 0;
  }

  isInSuccessState(resource: DaemonSet): boolean {
    return resource.podInfo.warnings.length === 0 && resource.podInfo.pending === 0;
  }

  hasErrors(daemonSet: DaemonSet): boolean {
    return daemonSet.podInfo.warnings.length > 0;
  }

  getEvents(daemonSet: DaemonSet): Event[] {
    return daemonSet.podInfo.warnings;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'pods', 'created', 'images'];
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
