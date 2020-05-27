

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../common/components/module';
import {SharedModule} from '../../shared.module';

import {DiscoveryComponent} from './component';
import {DiscoveryRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, DiscoveryRoutingModule],
  declarations: [DiscoveryComponent],
})
export class DiscoveryModule {}
