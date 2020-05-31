

import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {JobDetailComponent} from './detail/component';
import {JobListComponent} from './list/component';
import {JobRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, JobRoutingModule],
  declarations: [JobListComponent, JobDetailComponent],
})
export class JobModule {}
