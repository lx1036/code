

import {HttpParams} from '@angular/common/http';
import {ChangeDetectionStrategy, ChangeDetectorRef, Component, Input} from '@angular/core';
import {StorageClass, StorageClassList} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';

import {ResourceListBase} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-storage-class-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class StorageClassListComponent extends ResourceListBase<StorageClassList, StorageClass> {
  @Input() endpoint = EndpointManager.resource(Resource.storageClass).list();

  constructor(
    private readonly sc_: ResourceService<StorageClassList>,
    notifications: NotificationsService,
    cdr: ChangeDetectorRef,
  ) {
    super('storageclass', notifications, cdr);
    this.id = ListIdentifier.storageClass;
    this.groupId = ListGroupIdentifier.cluster;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);
  }

  getResourceObservable(params?: HttpParams): Observable<StorageClassList> {
    return this.sc_.get(this.endpoint, undefined, params);
  }

  map(storageClassList: StorageClassList): StorageClass[] {
    return storageClassList.storageClasses;
  }

  getDisplayColumns(): string[] {
    return ['name', 'provisioner', 'params', 'created'];
  }
}
