

import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';

import {SharedModule} from '../../../shared.module';
import {ReplicationControllerDetailComponent} from '../replicationcontroller/detail/component';
import {ReplicationControllerListComponent} from './list/component';
import {ReplicationControllerRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, ReplicationControllerRoutingModule],
  declarations: [ReplicationControllerListComponent, ReplicationControllerDetailComponent],
})
export class ReplicationControllerModule {}
