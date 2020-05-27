

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {ClusterComponent} from './component';

export const CLUSTER_ROUTE: Route = {
  path: '',
  component: ClusterComponent,
  data: {
    breadcrumb: 'Cluster',
    link: ['', 'cluster'],
  },
};

@NgModule({
  imports: [RouterModule.forChild([CLUSTER_ROUTE])],
  exports: [RouterModule],
})
export class ClusterRoutingModule {}
