

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {DEFAULT_ACTIONBAR} from '../../../common/components/actionbars/routing';

import {CONFIG_ROUTE} from '../routing';

import {SecretDetailComponent} from './detail/component';
import {SecretListComponent} from './list/component';

const SECRET_LIST_ROUTE: Route = {
  path: '',
  component: SecretListComponent,
  data: {
    breadcrumb: 'Secrets',
    parent: CONFIG_ROUTE,
  },
};

const SECRET_DETAIL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName',
  component: SecretDetailComponent,
  data: {
    breadcrumb: '{{ resourceName }}',
    parent: SECRET_LIST_ROUTE,
  },
};

@NgModule({
  imports: [RouterModule.forChild([SECRET_LIST_ROUTE, SECRET_DETAIL_ROUTE, DEFAULT_ACTIONBAR])],
  exports: [RouterModule],
})
export class SecretRoutingModule {}
