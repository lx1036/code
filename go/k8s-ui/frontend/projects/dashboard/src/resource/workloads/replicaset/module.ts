

import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';

import {SharedModule} from '../../../shared.module';
import {ReplicaSetDetailComponent} from './detail/component';
import {ReplicaSetListComponent} from './list/component';
import {ReplicaSetRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, ReplicaSetRoutingModule],
  declarations: [ReplicaSetListComponent, ReplicaSetDetailComponent],
})
export class ReplicaSetModule {}
