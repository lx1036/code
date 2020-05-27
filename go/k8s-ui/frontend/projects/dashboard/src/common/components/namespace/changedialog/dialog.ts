

import {Component, Inject} from '@angular/core';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material/dialog';

@Component({
  selector: 'kd-namespace-change-dialog',
  templateUrl: 'template.html',
})
export class NamespaceChangeDialog {
  namespace: string;
  newNamespace: string;

  constructor(
    public dialogRef: MatDialogRef<NamespaceChangeDialog>,
    @Inject(MAT_DIALOG_DATA) public data: {namespace: string; newNamespace: string},
  ) {
    this.namespace = data.namespace;
    this.newNamespace = data.newNamespace;
  }

  onNoClick(): void {
    this.dialogRef.close();
  }
}
