
import {NgModule} from '@angular/core';

import {SharedModule} from '../../shared.module';
import {ComponentsModule} from '../components/module';

// import {AlertDialog} from './alert/dialog';
// import {DeleteResourceDialog} from './deleteresource/dialog';
// import {LogsDownloadDialog} from './download/dialog';
// import {EditResourceDialog} from './editresource/dialog';
// import {ScaleResourceDialog} from './scaleresource/dialog';
// import {TriggerResourceDialog} from './triggerresource/dialog';

@NgModule({
  imports: [SharedModule, ComponentsModule],
  declarations: [
    // AlertDialog,
    // EditResourceDialog,
    // DeleteResourceDialog,
    // LogsDownloadDialog,
    // ScaleResourceDialog,
    // TriggerResourceDialog,
  ],
  exports: [
    // AlertDialog,
    // EditResourceDialog,
    // DeleteResourceDialog,
    // LogsDownloadDialog,
    // ScaleResourceDialog,
    // TriggerResourceDialog,
  ],
  entryComponents: [
    // AlertDialog,
    // EditResourceDialog,
    // DeleteResourceDialog,
    // LogsDownloadDialog,
    // ScaleResourceDialog,
    // TriggerResourceDialog,
  ],
})
export class DialogsModule {}
