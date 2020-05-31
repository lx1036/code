

import {Component, Inject} from '@angular/core';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material/dialog';

export interface AlertDialogConfig {
  title: string;
  message: string;
  confirmLabel: string;
}

@Component({
  selector: 'kd-alert-dialog',
  templateUrl: 'template.html',
})
export class AlertDialog {
  constructor(
    public dialogRef: MatDialogRef<AlertDialog>,
    @Inject(MAT_DIALOG_DATA) public data: AlertDialogConfig,
  ) {}

  onNoClick(): void {
    this.dialogRef.close();
  }
}
