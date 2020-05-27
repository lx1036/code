

import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NamespaceDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {ResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-namespace-detail',
  templateUrl: './template.html',
})
export class NamespaceDetailComponent implements OnInit, OnDestroy {
  private namespaceSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.namespace);
  namespace: NamespaceDetail;
  isInitialized = false;
  eventListEndpoint: string;

  constructor(
    private readonly namespace_: ResourceService<NamespaceDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;

    this.eventListEndpoint = this.endpoint_.child(resourceName, Resource.event);

    this.namespaceSubscription_ = this.namespace_
      .get(this.endpoint_.detail(), resourceName)
      .subscribe((d: NamespaceDetail) => {
        this.namespace = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Namespace', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.namespaceSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
