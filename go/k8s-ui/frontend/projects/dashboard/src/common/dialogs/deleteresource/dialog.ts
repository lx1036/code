

import {Component, Inject} from '@angular/core';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material/dialog';
import {ResourceMeta} from '../../services/global/actionbar';

@Component({
  selector: 'kd-delete-resource-dialog',
  templateUrl: 'template.html',
})
export class DeleteResourceDialog {
  constructor(
    public dialogRef: MatDialogRef<DeleteResourceDialog>,
    @Inject(MAT_DIALOG_DATA) public data: ResourceMeta,
  ) {}

  onNoClick(): void {
    this.dialogRef.close();
  }
}
