

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../common/components/module';
import {SharedModule} from '../shared.module';

import {CRDRoutingModule} from './routing';
import {CRDDetailComponent} from './detail/component';
import {CRDListComponent} from './list/component';
import {CRDObjectDetailComponent} from './crdobject/component';

@NgModule({
  imports: [SharedModule, ComponentsModule, CRDRoutingModule],
  declarations: [CRDListComponent, CRDDetailComponent, CRDObjectDetailComponent],
})
export class CrdModule {}
