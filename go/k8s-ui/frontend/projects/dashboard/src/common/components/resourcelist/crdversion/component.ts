

import {ChangeDetectionStrategy, Component, Input} from '@angular/core';
import {CRDVersion} from '@api/backendapi';
import {MatTableDataSource} from '@angular/material/table';

@Component({
  selector: 'kd-crd-versions-list',
  templateUrl: './template.html',
})
export class CRDVersionListComponent {
  @Input() versions: CRDVersion[];
  @Input() initialized: boolean;

  getDisplayColumns(): string[] {
    return ['name', 'served', 'storage'];
  }

  getDataSource(): MatTableDataSource<CRDVersion> {
    const tableData = new MatTableDataSource<CRDVersion>();
    tableData.data = this.versions;

    return tableData;
  }
}
