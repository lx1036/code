

import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {PersistentVolumeClaimDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-persistent-volume-claim-detail',
  templateUrl: './template.html',
})
export class PersistentVolumeClaimDetailComponent implements OnInit, OnDestroy {
  private persistentVolumeClaimSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.persistentVolumeClaim, true);
  persistentVolumeClaim: PersistentVolumeClaimDetail;
  isInitialized = false;

  constructor(
    private readonly persistentVolumeClaim_: NamespacedResourceService<PersistentVolumeClaimDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;
    const resourceNamespace = this.activatedRoute_.snapshot.params.resourceNamespace;

    this.persistentVolumeClaimSubscription_ = this.persistentVolumeClaim_
      .get(this.endpoint_.detail(), resourceName, resourceNamespace)
      .subscribe((d: PersistentVolumeClaimDetail) => {
        this.persistentVolumeClaim = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(
          new ResourceMeta('Persistent Volume Claim', d.objectMeta, d.typeMeta),
        );
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.persistentVolumeClaimSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }
}
