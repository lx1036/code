

import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {JobDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-job-detail',
  templateUrl: './template.html',
})
export class JobDetailComponent implements OnInit, OnDestroy {
  private jobSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.job, true);
  job: JobDetail;
  isInitialized = false;
  eventListEndpoint: string;
  podListEndpoint: string;

  constructor(
    private readonly job_: NamespacedResourceService<JobDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;
    const resourceNamespace = this.activatedRoute_.snapshot.params.resourceNamespace;

    this.eventListEndpoint = this.endpoint_.child(resourceName, Resource.event, resourceNamespace);
    this.podListEndpoint = this.endpoint_.child(resourceName, Resource.pod, resourceNamespace);

    this.jobSubscription_ = this.job_
      .get(this.endpoint_.detail(), resourceName, resourceNamespace)
      .subscribe((d: JobDetail) => {
        this.job = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Job', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.jobSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
