

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {PluginListComponent} from './list/component';
import {PluginDetailComponent} from './detail/component';

export const PLUGIN_LIST_ROUTE: Route = {
  path: '',
  component: PluginListComponent,
  data: {
    breadcrumb: 'Plugins',
  },
};

export const PLUGIN_DETAIL_ROUTE: Route = {
  path: ':pluginNamespace/:pluginName',
  component: PluginDetailComponent,
  data: {
    breadcrumb: '{{ pluginName }}',
    parent: PLUGIN_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([PLUGIN_LIST_ROUTE, PLUGIN_DETAIL_ROUTE])],
  exports: [RouterModule],
})
export class PluginsRoutingModule {}
