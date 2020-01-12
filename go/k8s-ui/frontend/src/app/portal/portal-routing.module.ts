import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {PortalComponent} from './portal.component';
import {AuthCheckGuard} from '../shared/auth-check-guard.service';
import {AppComponent} from './app.component';


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
