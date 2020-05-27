

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';

import {ErrorComponent} from './component';

const ERROR_ROUTE: Route = {
  path: '',
  component: ErrorComponent,
  data: {
    breadcrumb: 'Error',
  },
};

@NgModule({
  imports: [RouterModule.forChild([ERROR_ROUTE])],
  exports: [RouterModule],
})
export class ErrorRoutingModule {}
