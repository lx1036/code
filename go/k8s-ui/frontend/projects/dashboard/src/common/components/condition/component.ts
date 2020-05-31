

import {Component, Input, OnInit} from '@angular/core';
import {MatTableDataSource} from '@angular/material/table';
import {Condition} from 'typings/backendapi';

@Component({
  selector: 'kd-condition-list',
  templateUrl: './template.html',
})
export class ConditionListComponent implements OnInit {
  @Input() initialized: boolean;
  @Input() conditions: Condition[];
  @Input() showLastProbeTime = true;
  private columns = {
    0: 'type',
    1: 'status',
    2: 'lastProbeTime',
    3: 'lastTransitionTime',
    4: 'reason',
    5: 'message',
  };

  ngOnInit(): void {
    if (!this.showLastProbeTime) {
      delete this.columns[2];
    }
  }

  getConditionsColumns(): {[key: number]: string} {
    return this.columns;
  }

  getColumnKeys(): string[] {
    return Object.values(this.columns);
  }

  getDataSource(): MatTableDataSource<Condition> {
    const tableData = new MatTableDataSource<Condition>();
    tableData.data = this.conditions;

    return tableData;
  }
}
