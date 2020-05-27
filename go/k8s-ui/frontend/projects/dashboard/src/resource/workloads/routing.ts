

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {WorkloadsComponent} from './component';

export const WORKLOADS_ROUTE: Route = {
  path: '',
  component: WorkloadsComponent,
  data: {
    breadcrumb: 'Workloads',
    link: ['', 'workloads'],
  },
};

@NgModule({
  imports: [RouterModule.forChild([WORKLOADS_ROUTE])],
  exports: [RouterModule],
})
export class WorkloadsRoutingModule {}
