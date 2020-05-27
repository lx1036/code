

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {LOGS_SCALE_DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {WORKLOADS_ROUTE} from '../routing';

import {ReplicaSetDetailComponent} from './detail/component';
import {ReplicaSetListComponent} from './list/component';

const REPLICASET_LIST_ROUTE: Route = {
  path: '',
  component: ReplicaSetListComponent,
  data: {
    breadcrumb: 'Replica Sets',
    parent: WORKLOADS_ROUTE,
  },
};

export const REPLICASET_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: ReplicaSetDetailComponent,
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
export class ReplicaSetRoutingModule {}
