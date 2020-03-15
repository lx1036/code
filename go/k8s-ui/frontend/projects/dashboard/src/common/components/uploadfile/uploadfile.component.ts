import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {HTMLInputEvent, KdFile} from "../../../typings/frontend-api";
import {MatDialog} from "@angular/material/dialog";
import {AlertDialog, AlertDialogConfig} from "../dialog/dialog";

@Component({
  selector: 'kube-upload-file',
  template: `
    <div fxLayout="row" fxLayoutAlign="space-between start">
      <mat-form-field fxFlex>
        <input matInput title="fileInput" [placeholder]="label" [ngModel]="filename" [readonly]="true" (click)="fileInput.click()" />
      </mat-form-field>
      <button mat-button type="button" color="primary" (click)="fileInput.click()" class="kube-upload-button">
        <mat-icon>more_horiz</mat-icon>
      </button>
      <input hidden type="file" #fileInput (change)="onChange($event)" />
    </div>
  `,
  styleUrls: ['./styles.scss']
})
export class UploadFileComponent {
  @Input() label: string;
  @Output() onLoad = new EventEmitter<KdFile>();
  filename: string;

  constructor(private readonly matDialog: MatDialog) {
  }

  onChange(event: HTMLInputEvent) {
    if (event.target.files.length > 0) {
      this.readFile(event.target.files[0]);
    }
  }

  readFile(file: File) {
    this.filename = file.name;
    const reader = new FileReader();
    reader.onload = (event: ProgressEvent) => {
      const content = (event.target as FileReader).result;
      this.onLoad.emit({
        name: this.filename,
        content,
      } as KdFile)
    };

    if (file instanceof ArrayBuffer) {
      this.reportError('File Format Error', 'Specified file has the wrong format');
    } else {
      reader.readAsText(file)
    }
  }

  private reportError(title: string, message: string): void {
    const configData: AlertDialogConfig = {
      title,
      message,
      confirmLabel: 'OK',
    };
    this.matDialog.open(AlertDialog, {data: configData});
  }
}
