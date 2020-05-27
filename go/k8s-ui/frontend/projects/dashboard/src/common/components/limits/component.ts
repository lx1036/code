

import {Component, Input} from '@angular/core';
import {MatTableDataSource} from '@angular/material/table';
import {LimitRange} from 'typings/backendapi';

@Component({
  selector: 'kd-resource-limit-list',
  templateUrl: './template.html',
})
export class ResourceLimitListComponent {
  @Input() initialized: boolean;
  @Input() limits: LimitRange[];

  getColumnIds(): string[] {
    return ['name', 'type', 'default', 'request'];
  }

  getDataSource(): MatTableDataSource<LimitRange> {
    const tableData = new MatTableDataSource<LimitRange>();
    tableData.data = this.limits;
    return tableData;
  }
}
