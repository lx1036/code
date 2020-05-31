

import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';

import {ComponentsModule} from '../common/components/module';
import {SharedModule} from '../shared.module';

import {ShellComponent} from './component';
import {ShellRoutingModule} from './routing';

@NgModule({
  imports: [CommonModule, SharedModule, ComponentsModule, ShellRoutingModule],
  declarations: [ShellComponent],
})
export class ShellModule {}
