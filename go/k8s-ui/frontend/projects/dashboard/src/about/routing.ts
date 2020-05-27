

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {ActionbarComponent} from './actionbar/component';
import {AboutComponent} from './component';

export const ABOUT_ROUTE: Route = {
  path: '',
  component: AboutComponent,
  data: {
    breadcrumb: 'About',
  },
};

export const ACTIONBAR = {
  path: '',
  component: ActionbarComponent,
  outlet: 'actionbar',
};

@NgModule({
  imports: [RouterModule.forChild([ABOUT_ROUTE, ACTIONBAR])],
  exports: [RouterModule],
})
export class AboutRoutingModule {}
