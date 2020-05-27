

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';

import {SEARCH_BREADCRUMB_PLACEHOLDER} from '../common/components/breadcrumbs/component';
import {SearchGuard} from '../common/services/guard/search';

import {SearchComponent} from './component';

export const SEARCH_ROUTE: Route = {
  path: '',
  component: SearchComponent,
  canDeactivate: [SearchGuard],
  data: {
    breadcrumb: SEARCH_BREADCRUMB_PLACEHOLDER,
  },
};

@NgModule({
  imports: [RouterModule.forChild([SEARCH_ROUTE])],
  exports: [RouterModule],
})
export class SearchRoutingModule {}
