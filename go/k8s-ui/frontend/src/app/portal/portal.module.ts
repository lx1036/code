import {NgModule} from '@angular/core';

import {PortalComponent} from './portal.component';
import {AppComponent} from './app.component';
import {AppUserComponent} from './app-user.component';
import {ListAppUserComponent} from './list-app-user.component';
import {AuthCheckGuard} from '../shared/auth-check-guard.service';
import {AuthService} from '../shared/auth.service';
import {RouterModule, Routes} from '@angular/router';

const routes: Routes = [
  {
    path: 'portal/namespace/:nid',
    canActivateChild: [AuthCheckGuard],
    component: PortalComponent,
    children: [
      {
        path: 'app',
        component: AppComponent
      }
    ],
  }
];

@NgModule({
  imports: [
    RouterModule.forChild(routes)
  ],
  exports: [
    RouterModule
  ],
  declarations: [],
  providers: [
    AuthCheckGuard,
  ],
})
export class PortalRoutingModule {
}

@NgModule({
  imports: [
    PortalRoutingModule,
  ],
  exports: [
    PortalRoutingModule,
  ],
  declarations: [
    PortalComponent,
    AppComponent,
    AppUserComponent,
    ListAppUserComponent,
  ],
  providers: [
    AuthCheckGuard,
    AuthService,
  ],
})
export class PortalModule {
}
