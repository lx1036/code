

import {Component, Input} from '@angular/core';
import {Endpoint} from '@api/backendapi';

/**
 * Component definition object for the component that displays the endpoints which are accessible
 * from the outside of the cluster.
 */
@Component({
  selector: 'kd-external-endpoint',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class ExternalEndpointComponent {
  @Input() endpoints: Endpoint[];
}
