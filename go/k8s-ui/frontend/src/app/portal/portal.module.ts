import {NgModule} from '@angular/core';

import {PortalComponent} from './portal.component';
import {PortalRoutingModule} from './portal-routing.module';
import {AppComponent} from "./app.component";
import {AppUserComponent} from "./app-user.component";
import {ListAppUserComponent} from "./list-app-user.component";
import {AuthCheckGuard} from "../shared/auth-check-guard.service";
import {AuthService} from "../shared/auth.service";

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
