

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {CLUSTER_ROUTE} from '../routing';

import {NodeDetailComponent} from './detail/component';
import {NodeListComponent} from './list/component';

const NODE_LIST_ROUTE: Route = {
  path: '',
  component: NodeListComponent,
  data: {
    breadcrumb: 'Nodes',
    parent: CLUSTER_ROUTE,
  },
};

const NODE_DETAIL_ROUTE: Route = {
  path: ':resourceName',
  component: NodeDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: NODE_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([NODE_LIST_ROUTE, NODE_DETAIL_ROUTE, DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class NodeRoutingModule {}
