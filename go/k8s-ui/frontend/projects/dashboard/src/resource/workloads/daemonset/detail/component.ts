

import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {DaemonSetDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-daemon-set-detail',
  templateUrl: './template.html',
})
export class DaemonSetDetailComponent implements OnInit, OnDestroy {
  private daemonSetSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.daemonSet, true);
  daemonSet: DaemonSetDetail;
  isInitialized = false;
  eventListEndpoint: string;
  podListEndpoint: string;
  serviceListEndpoint: string;

  constructor(
    private readonly daemonSet_: NamespacedResourceService<DaemonSetDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;
    const resourceNamespace = this.activatedRoute_.snapshot.params.resourceNamespace;

    this.eventListEndpoint = this.endpoint_.child(resourceName, Resource.event, resourceNamespace);
    this.podListEndpoint = this.endpoint_.child(resourceName, Resource.pod, resourceNamespace);
    this.serviceListEndpoint = this.endpoint_.child(
      resourceName,
      Resource.service,
      resourceNamespace,
    );

    this.daemonSetSubscription_ = this.daemonSet_
      .get(this.endpoint_.detail(), resourceName, resourceNamespace)
      .subscribe((d: DaemonSetDetail) => {
        this.daemonSet = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Daemon Set', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.daemonSetSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
