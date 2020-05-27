

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../common/components/module';
import {SharedModule} from '../../shared.module';
import {ClusterComponent} from './component';
import {ClusterRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, ClusterRoutingModule],
  declarations: [ClusterComponent],
})
export class ClusterModule {}
