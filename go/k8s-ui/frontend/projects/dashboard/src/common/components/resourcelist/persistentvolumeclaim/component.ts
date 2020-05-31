

import {HttpParams} from '@angular/common/http';
import {ChangeDetectionStrategy, ChangeDetectorRef, Component, Input} from '@angular/core';
import {Observable} from 'rxjs/Observable';
import {PersistentVolumeClaim, PersistentVolumeClaimList} from 'typings/backendapi';

import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-persistent-volume-claim-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class PersistentVolumeClaimListComponent extends ResourceListWithStatuses<
  PersistentVolumeClaimList,
  PersistentVolumeClaim
> {
  @Input() endpoint = EndpointManager.resource(Resource.persistentVolumeClaim, true).list();

  constructor(
    private readonly persistentVolumeClaim_: NamespacedResourceService<PersistentVolumeClaimList>,
    notifications: NotificationsService,
    cdr: ChangeDetectorRef,
  ) {
    super('persistentvolumeclaim', notifications, cdr);
    this.id = ListIdentifier.persistentVolumeClaim;
    this.groupId = ListGroupIdentifier.config;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInBoundState);
    this.registerBinding(this.icon.timelapse, 'kd-muted', this.isInPendingState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInLostState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  isInBoundState(resource: PersistentVolumeClaim): boolean {
    return resource.status === 'Bound';
  }

  isInPendingState(resource: PersistentVolumeClaim): boolean {
    return resource.status === 'Pending';
  }

  isInLostState(resource: PersistentVolumeClaim): boolean {
    return resource.status === 'Lost';
  }

  getResourceObservable(params?: HttpParams): Observable<PersistentVolumeClaimList> {
    return this.persistentVolumeClaim_.get(this.endpoint, undefined, undefined, params);
  }

  map(persistentVolumeClaimList: PersistentVolumeClaimList): PersistentVolumeClaim[] {
    return persistentVolumeClaimList.items;
  }

  getDisplayColumns(): string[] {
    return [
      'statusicon',
      'name',
      'labels',
      'status',
      'volume',
      'capacity',
      'accmodes',
      'storagecl',
      'created',
    ];
  }

  getVolumeHref(persistentVolumeName: string): string {
    return this.kdState_.href('persistentvolume', persistentVolumeName);
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
