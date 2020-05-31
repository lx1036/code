

import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../../common/components/module';
import {SharedModule} from '../../../shared.module';
import {CreateFromFormComponent} from './component';
import {CreateNamespaceDialog} from './createnamespace/dialog';
import {CreateSecretDialog} from './createsecret/dialog';
import {DeployLabelComponent} from './deploylabel/component';
import {EnvironmentVariablesComponent} from './environmentvariables/component';
import {HelpSectionComponent} from './helpsection/component';
import {UserHelpComponent} from './helpsection/userhelp/component';
import {PortMappingsComponent} from './portmappings/component';
import {UniqueNameValidator} from './validator/uniquename.validator';
import {ValidImageReferenceValidator} from './validator/validimagereference.validator';
import {ProtocolValidator} from './validator/validprotocol.validator';
import {WarnThresholdValidator} from './validator/warnthreshold.validator';

@NgModule({
  declarations: [
    HelpSectionComponent,
    UserHelpComponent,
    CreateFromFormComponent,
    CreateNamespaceDialog,
    CreateSecretDialog,
    EnvironmentVariablesComponent,
    UniqueNameValidator,
    ValidImageReferenceValidator,
    PortMappingsComponent,
    ProtocolValidator,
    DeployLabelComponent,
    WarnThresholdValidator,
  ],
  imports: [CommonModule, SharedModule, ComponentsModule],
  exports: [CreateFromFormComponent],
  entryComponents: [CreateNamespaceDialog, CreateSecretDialog],
})
export class CreateFromFormModule {}
