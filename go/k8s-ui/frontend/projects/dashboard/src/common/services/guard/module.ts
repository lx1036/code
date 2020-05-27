

import {NgModule} from '@angular/core';
import {AuthGuard} from './auth';
import {LoginGuard} from './login';
import {SearchGuard} from './search';

@NgModule({
  providers: [AuthGuard, SearchGuard, LoginGuard],
})
export class GuardsModule {}
