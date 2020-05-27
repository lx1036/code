

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DiscoveryComponent} from './component';

export const DISCOVERY_ROUTE: Route = {
  path: '',
  component: DiscoveryComponent,
  data: {
    breadcrumb: 'Discovery and Load Balancing',
    link: ['', 'discovery'],
  },
};

@NgModule({
  imports: [RouterModule.forChild([DISCOVERY_ROUTE])],
  exports: [RouterModule],
})
export class DiscoveryRoutingModule {}
