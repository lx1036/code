

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../common/components/module';
import {SharedModule} from '../shared.module';

import {ActionbarComponent} from './actionbar/component';
import {AboutComponent} from './component';
import {AboutRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, AboutRoutingModule],
  declarations: [AboutComponent, ActionbarComponent],
})
export class AboutModule {}
