

import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';

import {SharedModule} from '../../../shared.module';
import {StatefulSetDetailComponent} from './detail/component';
import {StatefulSetListComponent} from './list/component';
import {StatefulSetRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, StatefulSetRoutingModule],
  declarations: [StatefulSetListComponent, StatefulSetDetailComponent],
})
export class StatefulSetModule {}
