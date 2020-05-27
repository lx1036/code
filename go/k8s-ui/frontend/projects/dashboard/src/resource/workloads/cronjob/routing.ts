

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {TRIGGER_DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {WORKLOADS_ROUTE} from '../routing';

import {CronJobDetailComponent} from './detail/component';
import {CronJobListComponent} from './list/component';

const CRONJOB_LIST_ROUTE: Route = {
  path: '',
  component: CronJobListComponent,
  data: {
    breadcrumb: 'Cron Jobs',
    parent: WORKLOADS_ROUTE,
  },
};

const CRONJOB_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: CronJobDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: CRONJOB_LIST_ROUTE,
  },
};

@NgModule({
  imports: [
    RouterModule.forChild([CRONJOB_LIST_ROUTE, CRONJOB_DETAIL_ROUTE, TRIGGER_DEFAULT_ACTIONBAR]),
  ],
  exports: [RouterModule],
})
export class CronJobRoutingModule {}
