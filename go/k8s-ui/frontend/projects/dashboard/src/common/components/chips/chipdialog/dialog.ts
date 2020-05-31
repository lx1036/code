

import {Component, Inject} from '@angular/core';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material/dialog';
import {Chip} from '../component';

@Component({
  selector: 'kd-chip-dialog',
  templateUrl: 'template.html',
})
export class ChipDialog {
  constructor(
    public dialogRef: MatDialogRef<ChipDialog>,
    @Inject(MAT_DIALOG_DATA) public data: Chip,
  ) {}

  onNoClick(): void {
    this.dialogRef.close();
  }
}
