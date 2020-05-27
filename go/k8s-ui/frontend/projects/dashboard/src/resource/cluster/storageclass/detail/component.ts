

import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {StorageClassDetail} from '@api/backendapi';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {ResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-storage-class-detail',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class StorageClassDetailComponent implements OnInit, OnDestroy {
  private storageClassSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.storageClass);
  storageClass: StorageClassDetail;
  pvListEndpoint: string;
  isInitialized = false;

  constructor(
    private readonly storageClass_: ResourceService<StorageClassDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
  ) {}

  ngOnInit(): void {
    const resourceName = this.activatedRoute_.snapshot.params.resourceName;

    this.pvListEndpoint = this.endpoint_.child(resourceName, Resource.persistentVolume);

    this.storageClassSubscription_ = this.storageClass_
      .get(this.endpoint_.detail(), resourceName)
      .subscribe((d: StorageClassDetail) => {
        this.storageClass = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta('Storage Class', d.objectMeta, d.typeMeta));
        this.isInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.storageClassSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }

  getParameterNames(): string[] {
    return !!this.storageClass.parameters ? Object.keys(this.storageClass.parameters) : [];
  }
}
