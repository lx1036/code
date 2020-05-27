

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../common/components/module';
import {SharedModule} from '../../shared.module';

import {WorkloadsComponent} from './component';
import {WorkloadsRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, WorkloadsRoutingModule],
  declarations: [WorkloadsComponent],
})
export class WorkloadsModule {}
