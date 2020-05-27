

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {OverviewComponent} from './component';

export const OVERVIEW_ROUTE: Route = {
  path: '',
  component: OverviewComponent,
  data: {
    breadcrumb: 'Overview',
  },
};

@NgModule({
  imports: [RouterModule.forChild([OVERVIEW_ROUTE])],
  exports: [RouterModule],
})
export class OverviewRoutingModule {}
