

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../common/components/module';
import {WorkloadStatusComponent} from '../common/components/workloadstatus/component';
import {SharedModule} from '../shared.module';

import {OverviewComponent} from './component';
import {OverviewRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, OverviewRoutingModule],
  declarations: [OverviewComponent],
})
export class OverviewModule {}
