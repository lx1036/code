

import {HttpParams} from '@angular/common/http';
import {ChangeDetectionStrategy, ChangeDetectorRef, Component, Input} from '@angular/core';
import {ClusterRole, ClusterRoleList} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';

import {ResourceListBase} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-cluster-role-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ClusterRoleListComponent extends ResourceListBase<ClusterRoleList, ClusterRole> {
  @Input() endpoint = EndpointManager.resource(Resource.clusterRole).list();

  constructor(
    private readonly clusterRole_: ResourceService<ClusterRoleList>,
    notifications: NotificationsService,
    cdr: ChangeDetectorRef,
  ) {
    super('clusterrole', notifications, cdr);
    this.id = ListIdentifier.clusterRole;
    this.groupId = ListGroupIdentifier.cluster;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);
  }

  getResourceObservable(params?: HttpParams): Observable<ClusterRoleList> {
    return this.clusterRole_.get(this.endpoint, undefined, params);
  }

  map(clusterRoleList: ClusterRoleList): ClusterRole[] {
    return clusterRoleList.items;
  }

  getDisplayColumns(): string[] {
    return ['name', 'created'];
  }
}
