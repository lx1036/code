import {NgModule} from '@angular/core';

import {PortalComponent} from './portal.component';
import {PortalRoutingModule} from "./portal-routing.module";
import { AppComponent } from './app/app.component';

@NgModule({
  imports: [
    PortalRoutingModule,
    
  ],
  exports: [],
  declarations: [PortalComponent, AppComponent],
  providers: [],
})
export class PortalModule {
}
