import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {PortalComponent} from './portal.component';
import {AppComponent} from './app/app.component';
import {AuthCheckGuard} from '../shared/auth/auth-check-guard.service';


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
  providers: [],
})
export class PortalRoutingModule {
}
