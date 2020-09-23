import {RouterModule, Routes} from "@angular/router";
import {AuthCheckGuard} from "../shared/auth-check-guard.service";
import {PortalComponent} from "./portal.component";
import {AppComponent} from "./app.component";
import {NamespaceReportComponent} from "./namespace-report.component";
import {NgModule} from "@angular/core";
import {BaseComponent} from "./base.component";
import {ServiceComponent} from "./service/service.component";


const routes: Routes = [
  {
    path: 'portal/namespace/:nid',
    canActivateChild: [AuthCheckGuard],
    component: PortalComponent,
    children: [
      {path: 'app', component: AppComponent},
      {path: 'overview', component: NamespaceReportComponent},
      {
        path: 'app/:id', component: BaseComponent,
        children: [
          {path: 'service', component: ServiceComponent},
          {path: 'service/:serviceId', component: ServiceComponent},
          {path: 'service/:serviceId/tpl', component: CreateEditServiceTplComponent},
          {path: 'service/:serviceId/tpl/:tplId', component: CreateEditServiceTplComponent},
        ]
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
