

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {LOGS_SCALE_DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {WORKLOADS_ROUTE} from '../routing';

import {ReplicationControllerDetailComponent} from './detail/component';
import {ReplicationControllerListComponent} from './list/component';

const REPLICATIONCONTROLLER_LIST_ROUTE: Route = {
  path: '',
  component: ReplicationControllerListComponent,
  data: {
    breadcrumb: 'Replication Controllers',
    parent: WORKLOADS_ROUTE,
  },
};

export const REPLICATIONCONTROLLER_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: ReplicationControllerDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: REPLICATIONCONTROLLER_LIST_ROUTE,
  },
};

@NgModule({
  imports: [
    RouterModule.forChild([
      REPLICATIONCONTROLLER_LIST_ROUTE,
      REPLICATIONCONTROLLER_DETAIL_ROUTE,
      LOGS_SCALE_DEFAULT_ACTIONBAR,
    ]),
  ],
  exports: [RouterModule],
})
export class ReplicationControllerRoutingModule {}
