

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {CLUSTER_ROUTE} from '../routing';

import {ClusterRoleDetailComponent} from './detail/component';
import {ClusterRoleListComponent} from './list/component';

const CLUSTERROLE_LIST_ROUTE: Route = {
  path: '',
  component: ClusterRoleListComponent,
  data: {
    breadcrumb: 'Cluster Roles',
    parent: CLUSTER_ROUTE,
  },
};

const CLUSTERROLE_DETAIL_ROUTE: Route = {
  path: ':resourceName',
  component: ClusterRoleDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: CLUSTERROLE_LIST_ROUTE,
  },
};

@NgModule({
  imports: [
    RouterModule.forChild([CLUSTERROLE_LIST_ROUTE, CLUSTERROLE_DETAIL_ROUTE, DEFAULT_ACTIONBAR]),
  ],
  exports: [RouterModule],
})
export class ClusterRoutingModule {}
