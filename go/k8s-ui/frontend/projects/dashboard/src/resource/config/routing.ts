

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {ConfigComponent} from './component';

export const CONFIG_ROUTE: Route = {
  path: '',
  component: ConfigComponent,
  data: {
    breadcrumb: 'Config and Storage',
    link: ['', 'config'],
  },
};

@NgModule({
  imports: [RouterModule.forChild([CONFIG_ROUTE])],
  exports: [RouterModule],
})
export class ConfigRoutingModule {}
