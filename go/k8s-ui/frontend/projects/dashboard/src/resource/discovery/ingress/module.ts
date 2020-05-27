

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {IngressDetailComponent} from './detail/component';
import {IngressListComponent} from './list/component';
import {IngressRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, IngressRoutingModule],
  declarations: [IngressListComponent, IngressDetailComponent],
})
export class IngressModule {}
