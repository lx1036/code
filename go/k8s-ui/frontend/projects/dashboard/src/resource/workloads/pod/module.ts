

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {PodDetailComponent} from './detail/component';
import {PodListComponent} from './list/component';
import {PodRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, PodRoutingModule],
  declarations: [PodListComponent, PodDetailComponent],
})
export class PodModule {}
