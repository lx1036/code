

import {HttpParams} from '@angular/common/http';
import {ChangeDetectionStrategy, ChangeDetectorRef, Component, Input} from '@angular/core';
import {Plugin, PluginList} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';
import {ResourceListBase} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

@Component({
  selector: 'kd-plugin-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class PluginListComponent extends ResourceListBase<PluginList, Plugin> {
  @Input() endpoint = EndpointManager.resource(Resource.plugin, true).list();

  constructor(
    private readonly plugin_: NamespacedResourceService<PluginList>,
    notifications: NotificationsService,
    cdr: ChangeDetectorRef,
  ) {
    super('plugin', notifications, cdr);
    this.id = ListIdentifier.plugin;
    this.groupId = ListGroupIdentifier.none;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  getResourceObservable(params?: HttpParams): Observable<PluginList> {
    return this.plugin_.get(this.endpoint, undefined, undefined, params);
  }

  map(pluginList: PluginList): Plugin[] {
    return pluginList.items;
  }

  getDisplayColumns(): string[] {
    return ['name', 'dependencies', 'created'];
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
