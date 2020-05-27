

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {DISCOVERY_ROUTE} from '../routing';

import {ServiceDetailComponent} from './detail/component';
import {ServiceListComponent} from './list/component';

const SERVICE_LIST_ROUTE: Route = {
  path: '',
  component: ServiceListComponent,
  data: {
    breadcrumb: 'Services',
    parent: DISCOVERY_ROUTE,
  },
};

const SERVICE_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: ServiceDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: SERVICE_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([SERVICE_LIST_ROUTE, SERVICE_DETAIL_ROUTE, DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class ServiceRoutingModule {}
