

import {NgModule} from '@angular/core';
import {ComponentsModule} from '../common/components/module';
import {SharedModule} from '../shared.module';
import {PluginHolderComponent} from './holder';
import {PluginListComponent} from './list/component';
import {PluginsRoutingModule} from './routing';
import {PluginComponent} from './component';
import {PluginDetailComponent} from './detail/component';

@NgModule({
  imports: [SharedModule, ComponentsModule, PluginsRoutingModule],
  declarations: [
    PluginListComponent,
    PluginDetailComponent,
    PluginHolderComponent,
    PluginComponent,
  ],
})
export class PluginModule {}
