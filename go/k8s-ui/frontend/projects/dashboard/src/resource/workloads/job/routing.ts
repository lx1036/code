

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {LOGS_DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';
import {WORKLOADS_ROUTE} from '../routing';

import {JobDetailComponent} from './detail/component';
import {JobListComponent} from './list/component';

const JOB_LIST_ROUTE: Route = {
  path: '',
  component: JobListComponent,
  data: {
    breadcrumb: 'Jobs',
    parent: WORKLOADS_ROUTE,
  },
};

const JOB_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: JobDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: JOB_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([JOB_LIST_ROUTE, JOB_DETAIL_ROUTE, LOGS_DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class JobRoutingModule {}
