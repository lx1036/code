

import {Component, Input} from '@angular/core';
import {Endpoint} from '@api/backendapi';

/**
 * Component definition object for the component that displays the endpoints which are accessible
 * only from the inside of the cluster.
 */
@Component({
  selector: 'kd-internal-endpoint',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class InternalEndpointComponent {
  @Input() endpoints: Endpoint[];
}
