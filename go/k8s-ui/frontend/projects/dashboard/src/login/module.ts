

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../common/components/module';
import {SharedModule} from '../shared.module';
import {LoginComponent} from './component';

@NgModule({
  declarations: [LoginComponent],
  imports: [SharedModule, ComponentsModule],
})
export class LoginModule {}
