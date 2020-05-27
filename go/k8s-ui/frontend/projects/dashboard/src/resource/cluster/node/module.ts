

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {NodeDetailComponent} from './detail/component';
import {NodeListComponent} from './list/component';
import {NodeRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, NodeRoutingModule],
  declarations: [NodeListComponent, NodeDetailComponent],
})
export class NodeModule {}
