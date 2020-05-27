

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {LOGS_SCALE_DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {WORKLOADS_ROUTE} from '../routing';

import {StatefulSetDetailComponent} from './detail/component';
import {StatefulSetListComponent} from './list/component';

const REPLICASET_LIST_ROUTE: Route = {
  path: '',
  component: StatefulSetListComponent,
  data: {
    breadcrumb: 'Stateful Sets',
    parent: WORKLOADS_ROUTE,
  },
};

const REPLICASET_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: StatefulSetDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: REPLICASET_LIST_ROUTE,
  },
};

@NgModule({
  imports: [
    RouterModule.forChild([
      REPLICASET_LIST_ROUTE,
      REPLICASET_DETAIL_ROUTE,
      LOGS_SCALE_DEFAULT_ACTIONBAR,
    ]),
  ],
  exports: [RouterModule],
})
export class StatefulSetRoutingModule {}
