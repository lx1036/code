

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';

import {EXEC_PARENT_PLACEHOLDER} from '../common/components/breadcrumbs/component';

import {ShellComponent} from './component';

export const SHELL_ROUTE: Route = {
  path: ':resourceNamespace/:resourceName/:containerName',
  component: ShellComponent,
  data: {
    breadcrumb: 'Shell',
    parent: EXEC_PARENT_PLACEHOLDER,
  },
};

export const SHELL_ROUTE_RAW: Route = {
  path: ':resourceNamespace/:resourceName',
  component: ShellComponent,
  data: {
    breadcrumb: 'Shell',
    parent: EXEC_PARENT_PLACEHOLDER,
  },
};

@NgModule({
  imports: [RouterModule.forChild([SHELL_ROUTE_RAW, SHELL_ROUTE])],
  exports: [RouterModule],
})
export class ShellRoutingModule {}
