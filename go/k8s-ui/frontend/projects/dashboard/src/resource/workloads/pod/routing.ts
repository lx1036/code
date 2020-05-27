

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {LOGS_EXEC_DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {WORKLOADS_ROUTE} from '../routing';

import {PodDetailComponent} from './detail/component';
import {PodListComponent} from './list/component';

const POD_LIST_ROUTE: Route = {
  path: '',
  component: PodListComponent,
  data: {
    breadcrumb: 'Pods',
    parent: WORKLOADS_ROUTE,
  },
};

export const POD_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: PodDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: POD_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([POD_LIST_ROUTE, POD_DETAIL_ROUTE, LOGS_EXEC_DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class PodRoutingModule {}
