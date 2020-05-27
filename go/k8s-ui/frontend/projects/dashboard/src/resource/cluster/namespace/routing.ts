

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';

import {CLUSTER_ROUTE} from '../routing';

import {ActionbarComponent} from './detail/actionbar/component';
import {NamespaceDetailComponent} from './detail/component';
import {NamespaceListComponent} from './list/component';

const NAMESPACE_LIST_ROUTE: Route = {
  path: '',
  component: NamespaceListComponent,
  data: {
    breadcrumb: 'Namespaces',
    parent: CLUSTER_ROUTE,
  },
};

const NAMESPACE_DETAIL_ROUTE: Route = {
  path: ':resourceName',
  component: NamespaceDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: NAMESPACE_LIST_ROUTE,
  },
};

export const ACTIONBAR = {
  path: '',
  component: ActionbarComponent,
  outlet: 'actionbar',
};

@NgModule({
  imports: [RouterModule.forChild([NAMESPACE_LIST_ROUTE, NAMESPACE_DETAIL_ROUTE, ACTIONBAR])],
  exports: [RouterModule],
})
export class NamespaceRoutingModule {}
