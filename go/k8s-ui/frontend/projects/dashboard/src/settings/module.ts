

import {NgModule} from '@angular/core';

import {ComponentsModule} from '../common/components/module';
import {SharedModule} from '../shared.module';

import {SettingsComponent} from './component';
import {SettingsEntryComponent} from './entry/component';
import {GlobalSettingsComponent} from './global/component';
import {SaveAnywayDialog} from './global/saveanywaysdialog/dialog';
import {LocalSettingsComponent} from './local/component';
import {SettingsRoutingModule} from './routing';

@NgModule({
  imports: [SharedModule, ComponentsModule, SettingsRoutingModule],
  declarations: [
    GlobalSettingsComponent,
    LocalSettingsComponent,
    SettingsComponent,
    SettingsEntryComponent,
    SaveAnywayDialog,
  ],
  entryComponents: [SaveAnywayDialog],
})
export class SettingsModule {}
