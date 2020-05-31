

import {HttpParams} from '@angular/common/http';
import {ChangeDetectionStrategy, ChangeDetectorRef, Component, Input} from '@angular/core';
import {Observable} from 'rxjs/Observable';
import {ConfigMap, ConfigMapList} from 'typings/backendapi';
import {ResourceListBase} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-config-map-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ConfigMapListComponent extends ResourceListBase<ConfigMapList, ConfigMap> {
  @Input() endpoint = EndpointManager.resource(Resource.configMap, true).list();

  constructor(
    private readonly configMap_: NamespacedResourceService<ConfigMapList>,
    notifications: NotificationsService,
    cdr: ChangeDetectorRef,
  ) {
    super('configmap', notifications, cdr);
    this.id = ListIdentifier.configMap;
    this.groupId = ListGroupIdentifier.config;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  getResourceObservable(params?: HttpParams): Observable<ConfigMapList> {
    return this.configMap_.get(this.endpoint, undefined, undefined, params);
  }

  map(configMapList: ConfigMapList): ConfigMap[] {
    return configMapList.items;
  }

  getDisplayColumns(): string[] {
    return ['name', 'labels', 'created'];
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
