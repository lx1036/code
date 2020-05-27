

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../common/components/module';
import {SharedModule} from '../../shared.module';

import {ConfigComponent} from './component';
import {ConfigRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, ConfigRoutingModule],
  declarations: [ConfigComponent],
})
export class ConfigModule {}
