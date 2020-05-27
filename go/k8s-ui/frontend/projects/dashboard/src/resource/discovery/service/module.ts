

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {ServiceDetailComponent} from './detail/component';
import {ServiceListComponent} from './list/component';
import {ServiceRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, ServiceRoutingModule],
  declarations: [ServiceListComponent, ServiceDetailComponent],
})
export class ServiceModule {}
