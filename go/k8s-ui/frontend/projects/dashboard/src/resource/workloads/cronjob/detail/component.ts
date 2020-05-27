

import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {CronJobDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-cron-job-detail',
  templateUrl: './template.html',
})
export class CronJobDetailComponent implements OnInit, OnDestroy {
  private cronJobSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.cronJob, true);
  cronJob: CronJobDetail;
  isInitialized = false;
  eventListEndpoint: string;
  activeJobsEndpoint: string;
  inactiveJobsEndpoint: string;

  constructor(
    private readonly cronJob_: NamespacedResourceService<CronJobDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;
    const resourceNamespace = this.activatedRoute_.snapshot.params.resourceNamespace;

    this.eventListEndpoint = this.endpoint_.child(resourceName, Resource.event, resourceNamespace);
    this.activeJobsEndpoint = this.endpoint_.child(resourceName, Resource.job, resourceNamespace);
    this.inactiveJobsEndpoint = this.activeJobsEndpoint + `?active=false`;

    this.cronJobSubscription_ = this.cronJob_
      .get(this.endpoint_.detail(), resourceName, resourceNamespace)
      .subscribe((d: CronJobDetail) => {
        this.cronJob = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Cron Job', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.cronJobSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
