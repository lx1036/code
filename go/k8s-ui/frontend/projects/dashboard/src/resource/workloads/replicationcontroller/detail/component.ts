

import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {ReplicationControllerDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-replication-controller-detail',
  templateUrl: './template.html',
})
export class ReplicationControllerDetailComponent implements OnInit, OnDestroy {
  private replicationControllerSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.replicationController, true);
  replicationController: ReplicationControllerDetail;
  isInitialized = false;
  eventListEndpoint: string;
  podListEndpoint: string;
  serviceListEndpoint: string;

  constructor(
    private readonly replicationController_: NamespacedResourceService<ReplicationControllerDetail>,
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

    this.replicationControllerSubscription_ = this.replicationController_
      .get(this.endpoint_.detail(), resourceName, resourceNamespace)
      .subscribe((d: ReplicationControllerDetail) => {
        this.replicationController = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(
          new ResourceMeta('Replication Controller', d.objectMeta, d.typeMeta),
        );
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.replicationControllerSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
