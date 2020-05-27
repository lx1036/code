

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {DeploymentDetailComponent} from './detail/component';
import {DeploymentListComponent} from './list/component';
import {DeploymentRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, DeploymentRoutingModule],
  declarations: [DeploymentListComponent, DeploymentDetailComponent],
})
export class DeploymentModule {}
