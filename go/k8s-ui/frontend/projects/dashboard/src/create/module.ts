

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../common/components/module';
import {CreateServiceModule} from '../common/services/create/module';
import {SharedModule} from '../shared.module';

import {CreateComponent} from './component';
import {CreateFromFileComponent} from './from/file/component';
import {CreateFromFormModule} from './from/form/module';
import {CreateFromInputComponent} from './from/input/component';
import {CreateRoutingModule} from './routing';

@NgModule({
  imports: [
    SharedModule,
    ComponentsModule,
    CreateFromFormModule,
    CreateServiceModule,
    CreateRoutingModule,
  ],
  declarations: [CreateComponent, CreateFromInputComponent, CreateFromFileComponent],
})
export class CreateModule {}
