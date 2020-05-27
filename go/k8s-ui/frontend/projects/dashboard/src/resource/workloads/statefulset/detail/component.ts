

import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {StatefulSetDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-stateful-set-detail',
  templateUrl: './template.html',
})
export class StatefulSetDetailComponent implements OnInit, OnDestroy {
  private statefulSetSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.statefulSet, true);
  statefulSet: StatefulSetDetail;
  isInitialized = false;
  podListEndpoint: string;
  eventListEndpoint: string;

  constructor(
    private readonly statefulSet_: NamespacedResourceService<StatefulSetDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;
    const resourceNamespace = this.activatedRoute_.snapshot.params.resourceNamespace;

    this.podListEndpoint = this.endpoint_.child(resourceName, Resource.pod, resourceNamespace);
    this.eventListEndpoint = this.endpoint_.child(resourceName, Resource.event, resourceNamespace);

    this.statefulSetSubscription_ = this.statefulSet_
      .get(this.endpoint_.detail(), resourceName, resourceNamespace)
      .subscribe((d: StatefulSetDetail) => {
        this.statefulSet = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Stateful Set', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.statefulSetSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
