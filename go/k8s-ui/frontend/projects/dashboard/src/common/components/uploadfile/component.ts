

import {Component, EventEmitter, Input, Output} from '@angular/core';
import {MatDialog} from '@angular/material/dialog';
import {HTMLInputEvent, KdFile} from '@api/frontendapi';
import {AlertDialog, AlertDialogConfig} from 'common/dialogs/alert/dialog';

@Component({
  selector: 'kd-upload-file',
  templateUrl: 'template.html',
  styleUrls: ['style.scss'],
})
export class UploadFileComponent {
  @Input() label: string;
  @Output() onLoad = new EventEmitter<KdFile>();
  filename: string;

  constructor(private readonly matDialog_: MatDialog) {}

  onChange(event: HTMLInputEvent): void {
    if (event.target.files.length > 0) {
      this.readFile(event.target.files[0]);
    }
  }

  readFile(file: File): void {
    this.filename = file.name;

    const reader = new FileReader();
    reader.onload = (event: ProgressEvent) => {
      const content = (event.target as FileReader).result;
      this.onLoad.emit({
        name: this.filename,
        content,
      } as KdFile);
    };

    if (file instanceof ArrayBuffer) {
      this.reportError('File Format Error', 'Specified file has the wrong format');
    } else {
      reader.readAsText(file);
    }
  }

  private reportError(title: string, message: string): void {
    const configData: AlertDialogConfig = {
      title,
      message,
      confirmLabel: 'OK',
    };
    this.matDialog_.open(AlertDialog, {data: configData});
  }
}
