import {NgModule} from '@angular/core';

import {PortalComponent} from './portal.component';
import {SharedModule} from '../shared/shared.module';
import {NavComponent} from './nav.component';
import {MarkdownModule} from 'ngx-markdown';
import {NamespaceReportComponent} from './namespace-report.component';
import {OverviewComponent} from './overview.component';
import {ResourceReportComponent} from './resource.component';
import {HistoryComponent} from './history.component';
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
    NavComponent,
    NamespaceReportComponent,
    OverviewComponent,
    ResourceReportComponent,
    HistoryComponent,
  ],
  providers: [
    AuthService,
  ],
})
export class PortalModule {
}
