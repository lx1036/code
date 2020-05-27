

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {LOGS_DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {WORKLOADS_ROUTE} from '../routing';

import {DaemonSetDetailComponent} from './detail/component';
import {DaemonSetListComponent} from './list/component';

const DAEMONSET_LIST_ROUTE: Route = {
  path: '',
  component: DaemonSetListComponent,
  data: {
    breadcrumb: 'Daemon Sets',
    parent: WORKLOADS_ROUTE,
  },
};

const DAEMONSET_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: DaemonSetDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: DAEMONSET_LIST_ROUTE,
  },
};

@NgModule({
  imports: [
    RouterModule.forChild([DAEMONSET_LIST_ROUTE, DAEMONSET_DETAIL_ROUTE, LOGS_DEFAULT_ACTIONBAR]),
  ],
  exports: [RouterModule],
})
export class DaemonSetRoutingModule {}
