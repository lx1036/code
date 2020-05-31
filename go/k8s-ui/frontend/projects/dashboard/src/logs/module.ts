

import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';

import {ComponentsModule} from '../common/components/module';
import {SharedModule} from '../shared.module';

import {LogsComponent} from './component';
import {LogsRoutingModule} from './routing';

@NgModule({
  imports: [CommonModule, SharedModule, ComponentsModule, LogsRoutingModule],
  declarations: [LogsComponent],
})
export class LogsModule {}
