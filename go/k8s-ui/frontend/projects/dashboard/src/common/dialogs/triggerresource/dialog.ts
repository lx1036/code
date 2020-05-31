

import {Component, Inject} from '@angular/core';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material/dialog';
import {ResourceMeta} from '../../services/global/actionbar';

@Component({
  selector: 'kd-trigger-resource-dialog',
  templateUrl: 'template.html',
})
export class TriggerResourceDialog {
  constructor(
    public dialogRef: MatDialogRef<TriggerResourceDialog>,
    @Inject(MAT_DIALOG_DATA) public data: ResourceMeta,
  ) {}

  onNoClick(): void {
    this.dialogRef.close();
  }
}
