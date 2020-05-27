

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../common/components/module';
import {GlobalServicesModule} from '../common/services/global/module';
import {GuardsModule} from '../common/services/guard/module';
import {CoreModule} from '../core.module';
import {SharedModule} from '../shared.module';

import {ChromeComponent} from './component';
import {NavModule} from './nav/module';
import {NotificationsComponent} from './notifications/component';
import {ChromeRoutingModule} from './routing';
import {SearchComponent} from './search/component';
import {UserPanelComponent} from './userpanel/component';

@NgModule({
  imports: [SharedModule, ComponentsModule, NavModule, ChromeRoutingModule, GuardsModule],
  declarations: [ChromeComponent, SearchComponent, NotificationsComponent, UserPanelComponent],
})
export class ChromeModule {}
