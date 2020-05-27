

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {ClusterRoleDetailComponent} from './detail/component';
import {ClusterRoleListComponent} from './list/component';
import {ClusterRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, ClusterRoutingModule],
  declarations: [ClusterRoleListComponent, ClusterRoleDetailComponent],
})
export class ClusterRoleModule {}
