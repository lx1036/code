import {NgModule} from '@angular/core';

import {PortalComponent} from './portal.component';
import {AppComponent} from './app.component';
import {AppUserComponent} from './app-user.component';
import {ListAppUserComponent} from './list-app-user.component';
import {AuthCheckGuard} from '../shared/auth-check-guard.service';
import {AuthService} from '../shared/auth.service';
import {RouterModule, Routes} from '@angular/router';
import {SharedModule} from '../shared/shared.module';
import {NavComponent} from './nav.component';
import {CommonModule} from '@angular/common';
import {TranslateModule} from '@ngx-translate/core';
import {MarkdownModule} from 'ngx-markdown';

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
    RouterModule.forChild(routes),
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
    CommonModule,
    PortalRoutingModule,
    SharedModule,
    TranslateModule,
    MarkdownModule.forRoot(),
  ],
  exports: [],
  declarations: [
    PortalComponent,
    AppComponent,
    AppUserComponent,
    ListAppUserComponent,
    NavComponent,
  ],
  providers: [
    AuthCheckGuard,
    AuthService,
  ],
})
export class PortalModule {
}
