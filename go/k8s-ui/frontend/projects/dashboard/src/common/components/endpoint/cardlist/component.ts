

import {Component, Input} from '@angular/core';
import {MatTableDataSource} from '@angular/material/table';
import {Endpoint} from '@api/backendapi';

@Component({
  selector: 'kd-endpoint-card-list',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class EndpointListComponent {
  @Input() initialized: boolean;
  @Input() endpoints: Endpoint[];

  getEndpointsColumns(): string[] {
    return ['Host', 'Ports (Name, Port, Protocol)', 'Node', 'Ready'];
  }

  getDataSource(): MatTableDataSource<Endpoint> {
    const tableData = new MatTableDataSource<Endpoint>();
    tableData.data = this.endpoints;

    return tableData;
  }
}
