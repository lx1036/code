import {Component, Inject, OnInit} from '@angular/core';
import {MAT_DIALOG_DATA, MatDialogRef} from "@angular/material/dialog";

export interface AlertDialogConfig {
  title: string;
  message: string;
  confirmLabel: string;
}

@Component({
  selector: 'kube-alert-dialog',
  template: `
    <h2 mat-dialog-title>{{data.title}}</h2>
    <mat-dialog-content class="kd-dialog-text">{{data.message}}</mat-dialog-content>
    <mat-dialog-actions>
      <button mat-button color="primary" [mat-dialog-close]="true">{{data.confirmLabel}}</button>
    </mat-dialog-actions>
  `
})
export class AlertDialog {
  constructor(public dialogRef: MatDialogRef<AlertDialog>, @Inject(MAT_DIALOG_DATA) public data: AlertDialogConfig,) {}

  onNoClick(): void {
    this.dialogRef.close();
  }
}
