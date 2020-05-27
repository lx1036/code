

import {Component, Input} from '@angular/core';
import {PodInfo} from '@api/backendapi';

@Component({
  selector: 'kd-pod-status-card',
  templateUrl: './template.html',
})
export class PodStatusCardComponent {
  @Input() podInfo: PodInfo;
  @Input() initialized: boolean;
}
