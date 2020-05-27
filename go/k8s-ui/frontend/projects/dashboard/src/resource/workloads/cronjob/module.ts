

import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';

import {SharedModule} from '../../../shared.module';
import {CronJobDetailComponent} from './detail/component';
import {CronJobListComponent} from './list/component';
import {CronJobRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, CronJobRoutingModule],
  declarations: [CronJobListComponent, CronJobDetailComponent],
})
export class CronJobModule {}
