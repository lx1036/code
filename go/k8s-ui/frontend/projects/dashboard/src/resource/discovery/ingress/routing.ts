

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {DISCOVERY_ROUTE} from '../routing';

import {IngressDetailComponent} from './detail/component';
import {IngressListComponent} from './list/component';

const INGRESS_LIST_ROUTE: Route = {
  path: '',
  component: IngressListComponent,
  data: {
    breadcrumb: 'Ingresses',
    parent: DISCOVERY_ROUTE,
  },
};

const INGRESS_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: IngressDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: INGRESS_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([INGRESS_LIST_ROUTE, INGRESS_DETAIL_ROUTE, DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class IngressRoutingModule {}
