

import 'rxjs/add/operator/startWith';

import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {IngressDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-ingress-detail',
  templateUrl: './template.html',
})
export class IngressDetailComponent implements OnInit, OnDestroy {
  private ingressSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.ingress, true);
  ingress: IngressDetail;
  isInitialized = false;

  constructor(
    private readonly ingress_: NamespacedResourceService<IngressDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;
    const resourceNamespace = this.activatedRoute_.snapshot.params.resourceNamespace;

    this.ingressSubscription_ = this.ingress_
      .get(this.endpoint_.detail(), resourceName, resourceNamespace)
      .subscribe((d: IngressDetail) => {
        this.ingress = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Ingress', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.ingressSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
