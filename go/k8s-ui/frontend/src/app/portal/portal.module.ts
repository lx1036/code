import {NgModule} from '@angular/core';

import {PortalComponent} from './portal.component';
import {AppComponent, CreateEditAppComponent} from './app.component';
import {AppUserComponent} from './app-user.component';
import {ListAppUserComponent} from './list-app-user.component';
import {AuthCheckGuard} from '../shared/auth-check-guard.service';
import {SharedModule} from '../shared/shared.module';
import {NavComponent} from './nav.component';
import {CommonModule} from '@angular/common';
import {MarkdownModule} from 'ngx-markdown';
import {NamespaceReportComponent} from './namespace-report.component';
import {SidenavNamespaceComponent} from './sidenav-namespace.component';
import {OverviewComponent} from './overview.component';
import {ResourceReportComponent} from './resource.component';
import {HistoryComponent} from './history.component';
import {ListAppsComponent} from './list-apps.component';
import {PortalRoutingModule} from "./portal-routing.module";
import {AuthService} from "../shared/components/auth/auth.service";


@NgModule({
  imports: [
    PortalRoutingModule,
    SharedModule,
    MarkdownModule.forRoot(),
  ],
  exports: [],
  declarations: [
    PortalComponent,
    AppComponent,
    AppUserComponent,
    ListAppUserComponent,
    NavComponent,
    NamespaceReportComponent,
    SidenavNamespaceComponent,
    OverviewComponent,
    ResourceReportComponent,
    HistoryComponent,
    ListAppsComponent,
    CreateEditAppComponent,
  ],
  providers: [
    AuthCheckGuard,
    AuthService,
  ],
})
export class PortalModule {
}
