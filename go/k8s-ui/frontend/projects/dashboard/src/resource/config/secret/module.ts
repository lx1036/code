

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';

import {SecretDetailComponent} from './detail/component';
import {SecretDetailEditComponent} from './detail/edit/component';
import {SecretListComponent} from './list/component';
import {SecretRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, SecretRoutingModule],
  declarations: [SecretListComponent, SecretDetailComponent, SecretDetailEditComponent],
})
export class SecretModule {}
