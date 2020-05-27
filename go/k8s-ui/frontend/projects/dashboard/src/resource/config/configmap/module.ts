

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {ConfigMapDetailComponent} from './detail/component';
import {ConfigMapListComponent} from './list/component';
import {ConfigMapRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, ConfigMapRoutingModule],
  declarations: [ConfigMapListComponent, ConfigMapDetailComponent],
})
export class ConfigMapModule {}
