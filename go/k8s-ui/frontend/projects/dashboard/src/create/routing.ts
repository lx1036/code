

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';

import {CreateComponent} from './component';

const CREATE_ROUTE: Route = {
  path: '',
  component: CreateComponent,
  data: {
    breadcrumb: 'Create',
  },
};

@NgModule({
  imports: [RouterModule.forChild([CREATE_ROUTE])],
  exports: [RouterModule],
})
export class CreateRoutingModule {}
