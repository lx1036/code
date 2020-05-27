

import {Component, Input} from '@angular/core';
import {MatTableDataSource} from '@angular/material/table';
import {ResourceQuotaDetail} from 'typings/backendapi';

@Component({
  selector: 'kd-resource-quota-list',
  templateUrl: './template.html',
})
export class ResourceQuotaListComponent {
  @Input() initialized: boolean;
  @Input() quotas: ResourceQuotaDetail[];

  getQuotaColumns(): string[] {
    return ['name', 'created', 'status'];
  }

  getDataSource(): MatTableDataSource<ResourceQuotaDetail> {
    const tableData = new MatTableDataSource<ResourceQuotaDetail>();
    tableData.data = this.quotas;
    return tableData;
  }
}
