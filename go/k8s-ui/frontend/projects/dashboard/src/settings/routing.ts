

import {NgModule} from '@angular/core';
import {Route, RouterModule} from '@angular/router';
import {SettingsComponent} from './component';

export const SETTINGS_ROUTE: Route = {
  path: '',
  component: SettingsComponent,
  data: {
    breadcrumb: 'Settings',
  },
};

@NgModule({
  imports: [RouterModule.forChild([SETTINGS_ROUTE])],
  exports: [RouterModule],
})
export class SettingsRoutingModule {}
