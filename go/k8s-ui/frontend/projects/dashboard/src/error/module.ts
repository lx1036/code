

import {ErrorHandler, NgModule} from '@angular/core';

import {ComponentsModule} from '../common/components/module';
import {NavServiceModule} from '../common/services/nav/module';
import {SharedModule} from '../shared.module';
import {ErrorComponent} from './component';
import {GlobalErrorHandler} from './handler';
import {ErrorRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, NavServiceModule, ErrorRoutingModule],
  declarations: [ErrorComponent],
})
export class ErrorModule {}
