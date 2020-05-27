

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {CONFIG_ROUTE} from '../routing';

import {ConfigMapDetailComponent} from './detail/component';
import {ConfigMapListComponent} from './list/component';

const CONFIGMAP_LIST_ROUTE: Route = {
  path: '',
  component: ConfigMapListComponent,
  data: {
    breadcrumb: 'Config Maps',
    parent: CONFIG_ROUTE,
  },
};

const CONFIGMAP_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: ConfigMapDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: CONFIGMAP_LIST_ROUTE,
  },
};

@NgModule({
  imports: [
    RouterModule.forChild([CONFIGMAP_LIST_ROUTE, CONFIGMAP_DETAIL_ROUTE, DEFAULT_ACTIONBAR]),
  ],
  exports: [RouterModule],
})
export class ConfigMapRoutingModule {}
