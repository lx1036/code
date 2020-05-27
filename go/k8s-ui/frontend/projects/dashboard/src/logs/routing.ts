

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';

import {LOGS_PARENT_PLACEHOLDER} from '../common/components/breadcrumbs/component';

import {LogsComponent} from './component';

export const LOGS_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName/:resourceType',
  component: LogsComponent,
  data: {
    breadcrumb: 'Logs',
    parent: LOGS_PARENT_PLACEHOLDER,
  },
};

@NgModule({
  imports: [RouterModule.forChild([LOGS_ROUTE])],
  exports: [RouterModule],
})
export class LogsRoutingModule {}
