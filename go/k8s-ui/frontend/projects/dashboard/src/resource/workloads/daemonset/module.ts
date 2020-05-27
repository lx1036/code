

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {ActionbarComponent} from './detail/actionbar/component';
import {DaemonSetDetailComponent} from './detail/component';
import {DaemonSetListComponent} from './list/component';
import {DaemonSetRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, DaemonSetRoutingModule],
  declarations: [DaemonSetListComponent, DaemonSetDetailComponent, ActionbarComponent],
})
export class DaemonSetModule {}
