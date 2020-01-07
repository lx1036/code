import {NgModule} from '@angular/core';

import {PortalComponent} from './portal.component';
import {PortalRoutingModule} from './portal-routing.module';
import {AppComponent} from "./app.component";
import {AppUserComponent} from "./app-user.component";
import {ListAppUserComponent} from "./list-app-user.component";

@NgModule({
  imports: [
    PortalRoutingModule,
  ],
  exports: [],
  declarations: [PortalComponent, AppComponent, AppUserComponent, ListAppUserComponent],
  providers: [],
})
export class PortalModule {
}
