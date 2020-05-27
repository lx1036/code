

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {SCALE_DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {WORKLOADS_ROUTE} from '../routing';

import {DeploymentDetailComponent} from './detail/component';
import {DeploymentListComponent} from './list/component';

const DEPLOYMENT_LIST_ROUTE: Route = {
  path: '',
  component: DeploymentListComponent,
  data: {
    breadcrumb: 'Deployments',
    parent: WORKLOADS_ROUTE,
  },
};

const DEPLOYMENT_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: DeploymentDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: DEPLOYMENT_LIST_ROUTE,
  },
};

@NgModule({
  imports: [
    RouterModule.forChild([
      DEPLOYMENT_LIST_ROUTE,
      DEPLOYMENT_DETAIL_ROUTE,
      SCALE_DEFAULT_ACTIONBAR,
    ]),
  ],
  exports: [RouterModule],
})
export class DeploymentRoutingModule {}
