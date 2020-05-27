

import {HttpClient} from '@angular/common/http';
import {Component, Inject, OnInit} from '@angular/core';
import {MAT_DIALOG_DATA, MatDialogRef} from '@angular/material/dialog';
import {ReplicaCounts} from '@api/backendapi';

import {ResourceMeta} from '../../services/global/actionbar';

@Component({
  selector: 'kd-delete-resource-dialog',
  templateUrl: 'template.html',
})
export class ScaleResourceDialog implements OnInit {
  actual = 0;
  desired = 0;

  constructor(
    public dialogRef: MatDialogRef<ScaleResourceDialog>,
    @Inject(MAT_DIALOG_DATA) public data: ResourceMeta,
    private readonly http_: HttpClient,
  ) {}

  ngOnInit(): void {
    const url =
      `api/v1/scale/${this.data.typeMeta.kind}` +
      (this.data.objectMeta.namespace ? `/${this.data.objectMeta.namespace}` : '') +
      `/${this.data.objectMeta.name}/`;

    this.http_
      .get<ReplicaCounts>(url)
      .toPromise()
      .then(rc => {
        this.actual = rc.actualReplicas;
        this.desired = rc.desiredReplicas;
      });
  }

  onNoClick(): void {
    this.dialogRef.close();
  }
}
