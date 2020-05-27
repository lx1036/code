

import {HttpParams} from '@angular/common/http';
import {ChangeDetectionStrategy, ChangeDetectorRef, Component, Input} from '@angular/core';
import {CronJob, CronJobList, Metric} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-cron-job-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class CronJobListComponent extends ResourceListWithStatuses<CronJobList, CronJob> {
  @Input() endpoint = EndpointManager.resource(Resource.cronJob, true).list();
  @Input() showMetrics = false;
  cumulativeMetrics: Metric[];

  constructor(
    private readonly cronJob_: NamespacedResourceService<CronJobList>,
    notifications: NotificationsService,
    cdr: ChangeDetectorRef,
  ) {
    super('cronjob', notifications, cdr);
    this.id = ListIdentifier.cronJob;
    this.groupId = ListGroupIdentifier.workloads;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  getResourceObservable(params?: HttpParams): Observable<CronJobList> {
    return this.cronJob_.get(this.endpoint, undefined, undefined, params);
  }

  map(cronJobList: CronJobList): CronJob[] {
    this.cumulativeMetrics = cronJobList.cumulativeMetrics;
    return cronJobList.items;
  }

  isInErrorState(resource: CronJob): boolean {
    return resource.suspend;
  }

  isInSuccessState(resource: CronJob): boolean {
    return !resource.suspend;
  }

  getDisplayColumns(): string[] {
    return [
      'statusicon',
      'name',
      'labels',
      'schedule',
      'suspend',
      'active',
      'lastschedule',
      'created',
    ];
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
