import {RouterModule, Routes} from "@angular/router";
import {PortalComponent} from "./portal.component";
import {NamespaceReportComponent} from "./namespace-report.component";
import {NgModule} from "@angular/core";
import {BaseComponent} from "./base.component";
import {ServiceComponent} from "./service/service.component";
import {AuthCheckGuard} from "../shared/components/auth/auth-check-guard.service";
import {AppComponent} from "../app.component";
import {NamespaceApiKeyComponent} from "./namespace-apikey/apikey.component";


const routes: Routes = [
  {
    path: 'portal/namespace/:nid',
    canActivateChild: [AuthCheckGuard],
    component: PortalComponent,
    children: [
      {path: 'app', component: AppComponent},
      {path: 'apikey', component: NamespaceApiKeyComponent},
      {path: 'users', component: NamespaceUserComponent},
      {path: 'webhook', component: NamespaceWebHookComponent},
      {path: 'overview', component: NamespaceReportComponent},
      {
        path: 'app/:id', component: BaseComponent,
        children: [
          {path: 'service', component: ServiceComponent},
          {path: 'service/:serviceId', component: ServiceComponent},
          // {path: 'service/:serviceId/tpl', component: CreateEditServiceTplComponent},
          // {path: 'service/:serviceId/tpl/:tplId', component: CreateEditServiceTplComponent},
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
