

import {Component} from '@angular/core';
import {MatDialogRef} from '@angular/material/dialog';

@Component({
  selector: 'kd-settings-save-anyway-dialog',
  templateUrl: 'template.html',
})
export class SaveAnywayDialog {
  constructor(public dialogRef: MatDialogRef<SaveAnywayDialog>) {}

  onNoClick(): void {
    this.dialogRef.close();
  }
}
