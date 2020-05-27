

import {Routes} from '@angular/router';
import {LoginGuard} from './common/services/guard/login';
import {LoginComponent} from './login/component';

export const routes: Routes = [
  {path: 'login', component: LoginComponent, canActivate: [LoginGuard]},
  {path: '', redirectTo: '/overview', pathMatch: 'full'},
  {path: '**', redirectTo: '/overview'},
];
